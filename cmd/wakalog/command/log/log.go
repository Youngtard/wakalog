package log

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"github.com/Youngtard/wakalog/pkg/cmdutil"
	"github.com/Youngtard/wakalog/pkg/interact"
	wakasheets "github.com/Youngtard/wakalog/sheets"
	"github.com/Youngtard/wakalog/wakalog"
	"github.com/Youngtard/wakalog/wakatime"
	"github.com/icza/gox/timex"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"google.golang.org/api/sheets/v4"
)

type spreadsheet string
type week string

func (c spreadsheet) String() string {

	return string(c)

}

func (c week) String() string {

	return string(c)

}

func NewLogCommand(app *wakalog.Application) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "log",
		Short: "Log your summary activity",
		Long:  "Log your weekly summary activity to a Spreadsheet",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var wakatimeToken string
			var sheetsToken *oauth2.Token

			wakatimeToken, err := checkForWakaTimeToken()

			if err != nil {
				if errors.Is(err, wakalog.ErrWakaTimeTokenNotFound) {

					value, err := cmdutil.PromptForConfirmation(cmd.Context(), "You need to authenticate your WakaTime account in order to proceed. Proceed with authentication?")

					if err != nil {
						return err
					}

					if value {
						wakatimeToken, err = wakatime.Authorize(ctx)

						if err != nil {
							return &wakalog.AuthError{Err: fmt.Errorf("error authenticating with WakaTime")}
						}

					} else {
						return &wakalog.AuthError{Err: fmt.Errorf("unable to log. Authentication with WakaTime is required")}
					}

				} else {
					return err
				}
			}

			sheetsToken, err = checkForSheetsToken()

			if err != nil {
				if errors.Is(err, wakalog.ErrSheetsTokenNotFound) {

					value, err := cmdutil.PromptForConfirmation(cmd.Context(), "You need to authenticate your Google Sheets account in order to proceed. Proceed with authentication?")

					if err != nil {
						return err
					}

					if value {
						sheetsToken, err = wakasheets.Authorize(ctx)

						if err != nil {
							return &wakalog.AuthError{Err: fmt.Errorf("error authenticating with Google Sheets")}
						}
					} else {
						return &wakalog.AuthError{Err: fmt.Errorf("unable to log. Authentication with Google Sheets is required")}
					}

				} else {
					return err
				}
			}

			// TODO after 7 days / tokenExpiry + 7 days
			if time.Now().After(sheetsToken.Expiry) {
				sheetsToken, err = wakasheets.Authorize(ctx)

				if err != nil {

					return &wakalog.AuthError{Err: fmt.Errorf("error authenticating with Google Sheets")}

				}

			}

			app.InitializeWakaTime(wakatimeToken)

			err = app.InitializeSheets(ctx, sheetsToken)

			if err != nil {
				return fmt.Errorf("error initializing sheets service %w", err)
			}

			return nil

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			var selectedSheet spreadsheet
			var selectedWeek week

			// Continue as <name>?
			err := interact.TextInput("Enter your name (as seen on the Google Sheets document)", &name)

			if err != nil {
				return fmt.Errorf("error generating text input")
			}

			sheetsService := app.Sheets

			ssheet, err := sheetsService.Spreadsheets.Get(wakasheets.SpreadsheetId).Do()

			if err != nil {
				// TODO test errors
				return err
			}

			var spreadsheets []spreadsheet // spreadsheets representing months of the year
			for _, s := range ssheet.Sheets {

				spreadsheets = append(spreadsheets, spreadsheet(s.Properties.Title))
			}

			err = interact.Choice("Select a sheet to update", spreadsheets, &selectedSheet)

			if err != nil {
				return fmt.Errorf("error generating sheet selection")
			}

			var weeks []week

			weekRanges := []string{"C1:F1", "G1:J1", "K1:N1", "O1:R1"}

			for _, v := range weekRanges {
				weekRange := fmt.Sprintf("%s!%s", selectedSheet, v)

				resp, err := sheetsService.Spreadsheets.Values.Get(wakasheets.SpreadsheetId, weekRange).MajorDimension("ROWS").Do()

				if err != nil {
					// TODO test errors
					return err
				}

				for _, row := range resp.Values {

					for _, v := range row {

						weeks = append(weeks, week(v.(string)))
					}
				}

			}

			if len(weeks) == 0 {
				fmt.Println("No data found.")
				return nil
			}

			fmt.Println()

			err = interact.Choice("Select a week to update", weeks, &selectedWeek)

			if err != nil {
				return fmt.Errorf("error generating week selection")
			}

			return updateSheet(cmd.Context(), app, string(selectedSheet))

		},
	}

	return cmd
}

func checkForWakaTimeToken() (string, error) {

	var token string

	if err := wakatime.GetAccessToken(&token); err != nil {

		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", wakalog.ErrWakaTimeTokenNotFound
		}

		return "", wakalog.ErrGeneric

	} else {

		if len(strings.TrimSpace(token)) == 0 {

			return "", wakalog.ErrWakaTimeTokenNotFound

		}
	}

	return token, nil

}

func checkForSheetsToken() (*oauth2.Token, error) {
	token, err := wakasheets.RetrieveTokenFromFile()

	if err != nil {
		return nil, wakalog.ErrSheetsTokenNotFound
	}

	return token, nil
}

func updateSheet(ctx context.Context, app *wakalog.Application, sheet string) error {

	now := time.Now()

	currentYear, currentWeek := now.ISOWeek()

	dayOfWeek := int(now.Weekday())

	var relevantWeekOffset int

	/// Relevant days to work with is working days of the week (Mon - Fri)
	/// If it's weekend (Saturday, or Sunday), work with current week (Mon - Fri stats is available)
	/// Else, 1 week needs to be offset from current week in order to work with last week's data

	if dayOfWeek == 6 || dayOfWeek == 0 {
		relevantWeekOffset = 0

	} else {
		relevantWeekOffset = 1
	}

	relevantWeek := currentWeek - relevantWeekOffset

	startDate := timex.WeekStart(currentYear, relevantWeek)
	endDate := startDate.AddDate(0, 0, 4)

	summaries, err := app.WakaTime.GetSummaries(ctx, startDate, endDate)

	// todo remove
	fmt.Println(startDate)
	fmt.Println(endDate)

	if err != nil {
		return fmt.Errorf("error getting summaries: %w", err)
	}

	var rowIndex int

	firstDayOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 1, 1, 1, 1, time.UTC)
	_, firstDayOfCurrentMonthWeek := firstDayOfCurrentMonth.ISOWeek()

	weekInCurrentMonth := relevantWeek - firstDayOfCurrentMonthWeek

	startColumns := []string{"C", "G", "K", "O", "S"}
	startColumn := startColumns[weekInCurrentMonth-1]

	writeRange := fmt.Sprintf("%s!%s%d", sheet, startColumn, rowIndex)

	valuesRequest := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "RAW",
	}

	var valueRange sheets.ValueRange

	data := []interface{}{summaries.DailyAverage.Text, "", summaries.CumulativeTotal.Text}
	valueRange.Values = append(valueRange.Values, data)
	valueRange.Range = writeRange

	valuesRequest.Data = append(valuesRequest.Data, &valueRange)

	_, err = app.Sheets.Spreadsheets.Values.BatchUpdate(wakasheets.SpreadsheetId, valuesRequest).Do()

	if err != nil {
		return fmt.Errorf("unable to write data on sheet: %w", err)
	}

	return nil
}
