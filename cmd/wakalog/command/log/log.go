package log

import (
	"context"
	"errors"
	"fmt"

	"slices"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"github.com/Youngtard/wakalog/pkg/cmdutil"
	"github.com/Youngtard/wakalog/pkg/interact"
	wakasheets "github.com/Youngtard/wakalog/sheets"
	"github.com/Youngtard/wakalog/wakalog"
	"github.com/Youngtard/wakalog/wakatime"
	"github.com/charmbracelet/huh"
	"github.com/icza/gox/timex"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"google.golang.org/api/sheets/v4"
)

func NewLogCommand(app *wakalog.Application) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "log",
		Short: "Log your summary activity",
		Long:  "Log your weekly summary activity to a Spreadsheet",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var wakatimeAPIKey string
			var sheetsToken *oauth2.Token

			wakatimeAPIKey, err := checkForWakaTimeAPIKey()

			if err != nil {
				if errors.Is(err, wakalog.ErrWakaTimeAPIKeyNotFound) {

					wakatimeAPIKey, err = wakatime.Authorize(ctx)

					if err != nil {
						return &wakalog.AuthError{Err: fmt.Errorf("error authenticating with WakaTime")}
					}

				} else {
					return fmt.Errorf("error checking for wakatime api key: %w", err)
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

			app.InitializeWakaTime(wakatimeAPIKey)

			err = app.InitializeSheets(ctx, sheetsToken)

			if err != nil {
				return fmt.Errorf("error initializing sheets service %w", err)
			}

			return nil

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			var relevantSheet string

			sheetsService := app.Sheets

			ssheet, err := sheetsService.Spreadsheets.Get(wakasheets.SpreadsheetId).Do()

			if err != nil {
				// TODO test errors and other errors
				return err
			}

			startDate, _ := getRelevantStartAndEndDate()
			relevantMonth := startDate.Month()

			for i, s := range ssheet.Sheets {

				if i == int(relevantMonth)-1 {
					relevantSheet = s.Properties.Title
					break
				}

			}

			// TODO Continue as <name>?
			// TODO handle empty name
			err = interact.TextInput("Enter your name (as seen on the Google Sheets document - case sensitive)", &name)

			if err != nil {
				return fmt.Errorf("error generating name input")
			}

			/// Fetch names on sheet
			namesRange := fmt.Sprintf("%s!B3:B", relevantSheet)
			resp, err := sheetsService.Spreadsheets.Values.Get(wakasheets.SpreadsheetId, namesRange).MajorDimension("COLUMNS").Do()

			if err != nil {
				return fmt.Errorf("error retrieving names on sheet")
			}

			var rowIndex int
			var nameFound bool

			/// Check if inputted name exists on sheet, if so, store row index
			if len(resp.Values) == 0 {
				fmt.Println("No data found.")
				// TODO
				return nil
			} else {
				for _, row := range resp.Values {
					for i, v := range row {

						v := v.(string)

						if strings.TrimSpace(v) == strings.TrimSpace(name) {
							nameFound = true
							rowIndex = i + 3 // row cells start counting from 1 (so +1), then ignore first two rows (they are headers/don't contain user's data)
						}
					}
				}
			}

			if !nameFound {
				// TODO ask for name again
				fmt.Println("Name not found on sheet")
				return nil
			}

			err = updateSheet(cmd.Context(), app, relevantSheet, rowIndex)

			if err != nil {
				return fmt.Errorf("error updating sheet: %w", err)
			}

			fmt.Println("Sheet updated successfully :)")

			return nil

		},
	}

	return cmd
}

func checkForWakaTimeAPIKey() (string, error) {

	var apiKey string

	if err := wakatime.GetAPIKey(&apiKey); err != nil {

		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", wakalog.ErrWakaTimeAPIKeyNotFound
		}

		return "", wakalog.ErrGeneric

	} else {

		if len(strings.TrimSpace(apiKey)) == 0 {

			return "", wakalog.ErrWakaTimeAPIKeyNotFound

		}
	}

	return apiKey, nil

}

func checkForSheetsToken() (*oauth2.Token, error) {
	token, err := wakasheets.RetrieveTokenFromFile()

	if err != nil {
		return nil, wakalog.ErrSheetsTokenNotFound
	}

	return token, nil
}

func getRelevantStartAndEndDate() (time.Time, time.Time) {

	now := time.Now()

	currentYear, currentWeek := now.ISOWeek()

	dayOfWeek := int(now.Weekday())

	var relevantWeekOffset int

	/// Relevant days to work with is working days of the week (Mon - Fri)
	/// If it's weekend (Saturday, or Sunday), work with current week (Mon - Fri stats is available)
	/// Else, 1 week needs to be offset from current week in order to work with last week's data as the current working week is not yet over

	if dayOfWeek == 6 || dayOfWeek == 0 {
		relevantWeekOffset = 0

	} else {
		relevantWeekOffset = 1
	}

	relevantWeek := currentWeek - relevantWeekOffset

	startDate := timex.WeekStart(currentYear, relevantWeek)
	endDate := startDate.AddDate(0, 0, 4)

	return startDate, endDate

}

func updateSheet(ctx context.Context, app *wakalog.Application, sheet string, rowIndex int) error {

	startDate, endDate := getRelevantStartAndEndDate()

	summaries, err := app.WakaTime.GetSummaries(ctx, startDate, endDate)

	if err != nil {
		return fmt.Errorf("error getting summaries: %w", err)
	}

	var projectOptions []huh.Option[string]
	var projects []string
	var selectedProjects []string

	// Loop over period/days e.g. Mon-Fri
	for _, data := range summaries.Data {

		for _, project := range data.Projects {

			name := project.Name

			// Get unique list of projects worked on during period
			if !slices.Contains(projects, name) {
				projects = append(projects, name)
				projectOptions = append(projectOptions, huh.NewOption(name, name))
			}

		}
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select projects to get weekly activity from").
				Options(
					projectOptions...,
				).
				Value(&selectedProjects),
		),
	)

	err = form.RunWithContext(ctx)

	if err != nil {
		return fmt.Errorf("error generating project options")
	}

	if len(selectedProjects) == 0 {
		fmt.Println("No project selected. All projects will be used")
		selectedProjects = projects
	}

	startColumns := []string{"C", "G", "K", "O", "S"} // representing 5 possible weeks in a month

	relevantWeek := startDate.Day() / 7

	startColumn := startColumns[relevantWeek]

	writeRange := fmt.Sprintf("%s!%s%d", sheet, startColumn, rowIndex)

	valuesRequest := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "RAW",
	}

	var valueRange sheets.ValueRange

	var daysWorked int
	var totalTimePerDay []time.Duration

	// Loop over period/days e.g. Mon-Fri
	for _, data := range summaries.Data {

		var totalTimeForDay time.Duration

		for _, project := range data.Projects {

			projectName := project.Name

			if slices.Contains(selectedProjects, projectName) {

				totalTime, _ := time.ParseDuration(fmt.Sprintf("%dh%dm%ds", project.Hours, project.Minutes, project.Seconds))

				totalTimeForDay += totalTime

			}

		}

		totalTimePerDay = append(totalTimePerDay, totalTimeForDay)
	}

	var cummulativeTotalTime time.Duration

	for _, totalTime := range totalTimePerDay {

		// Skip days where user had no coding activity.
		// Daily Average computation according to WakaTime ignores days of no activity
		if totalTime <= time.Duration(0) {
			break
		}

		daysWorked += 1
		cummulativeTotalTime += totalTime

	}

	dailyAverage := cummulativeTotalTime.Hours() / float64(daysWorked)

	data := []interface{}{time.Duration(dailyAverage * float64(time.Hour)).String(), "", cummulativeTotalTime.String()}
	valueRange.Values = append(valueRange.Values, data)
	valueRange.Range = writeRange

	valuesRequest.Data = append(valuesRequest.Data, &valueRange)

	_, err = app.Sheets.Spreadsheets.Values.BatchUpdate(wakasheets.SpreadsheetId, valuesRequest).Do()

	if err != nil {
		return fmt.Errorf("unable to write data on sheet: %w", err)
	}

	return nil
}
