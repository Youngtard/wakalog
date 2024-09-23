package log

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"slices"
	"strings"
	"time"

	wakasheets "github.com/Youngtard/wakalog/sheets"
	"github.com/Youngtard/wakalog/wakalog"
	"github.com/Youngtard/wakalog/wakatime"
	"github.com/charmbracelet/huh"
	"github.com/icza/gox/timex"
	"github.com/savioxavier/termlink"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"google.golang.org/api/sheets/v4"
)

var errNoProjects = errors.New("no projects")

func NewLogCommand(app *wakalog.Application) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "log",
		Short: "Log your summary activity",
		Long:  "Log your weekly summary activity to a Spreadsheet",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var wakatimeAPIKey string
			var sheetsClient *http.Client

			wakatimeAPIKey, err := checkForWakaTimeAPIKey()

			if err != nil {
				if errors.Is(err, wakalog.ErrWakaTimeAPIKeyNotFound) {

					wakatimeAPIKey, err = wakatime.Authorize(ctx)

					if err != nil {
						return &wakalog.AuthError{Err: fmt.Errorf("error authenticating with WakaTime: %w", err)}
					}

				} else {
					return fmt.Errorf("error checking for wakatime api key: %w", err)
				}
			}

			sheetsClient, err = wakasheets.GetClient(ctx)

			if err != nil {
				return fmt.Errorf("error getting google client: %w", err)
			}

			app.InitializeWakaTime(wakatimeAPIKey)

			err = app.InitializeSheets(ctx, sheetsClient)

			if err != nil {
				return fmt.Errorf("error initializing sheets service: %w", err)
			}

			return nil

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var username string
			var relevantSheet string
			var relevantSheetId int64

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

					relevantSheetId = s.Properties.SheetId
					relevantSheet = s.Properties.Title
					break
				}

			}

			/// Fetch names on sheet
			namesRange := fmt.Sprintf("%s!B3:B", relevantSheet)
			resp, err := sheetsService.Spreadsheets.Values.Get(wakasheets.SpreadsheetId, namesRange).MajorDimension("COLUMNS").Do()

			if err != nil {
				return fmt.Errorf("error retrieving usernames on sheet: %w", err)
			}

			var namesOnSheet []string

			if len(resp.Values) == 0 {
				fmt.Println("No username data found.")
				return nil
			} else {
				for _, row := range resp.Values {
					for _, v := range row {

						v := v.(string)

						name := strings.TrimSpace(v)

						namesOnSheet = append(namesOnSheet, name)

					}
				}
			}

			// TODO Continue as <name>?
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Enter your name (as seen on the Google Sheets document - case sensitive)").
						Placeholder("Enter Name...").
						Value(&username).
						Suggestions(namesOnSheet).
						Validate(func(value string) error {
							if len(value) == 0 {
								return fmt.Errorf("Your name is required to proceed.")
							}

							if !slices.Contains(namesOnSheet, value) {
								return fmt.Errorf("Name not found on sheet.")
							}
							return nil
						}).WithTheme(huh.ThemeBase()),
				),
			)

			err = form.RunWithContext(ctx)

			if err != nil {
				return fmt.Errorf("error getting username: %w", err)
			}

			var rowIndex int

			for i, name := range namesOnSheet {
				if name == username {
					rowIndex = i + 3 // row cells start counting from 1 (so +1), then ignore first two rows (they are headers/don't contain user's data)
					break
				}
			}

			err = updateSheet(ctx, app, relevantSheet, rowIndex, relevantSheetId)

			if err != nil {

				if errors.Is(err, errNoProjects) {
					fmt.Println("No projects data found for period. Don't have WakaTime? Install WakaTime plugin on your IDE to get started.")
					return nil
				}
				return fmt.Errorf("error updating sheet: %w", err)
			}

			linkToSheet := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/edit?gid=%d#gid=%d", wakasheets.SpreadsheetId, relevantSheetId, relevantSheetId)

			fmt.Printf("Sheet updated successfully :)\nView sheet %s.\n", termlink.ColorLink("here", linkToSheet, "blue"))

			return nil

		},
	}

	return cmd
}

func checkForWakaTimeAPIKey() (string, error) {

	var apiKey string

	if err := wakatime.GetAPIKey(&apiKey); err != nil {

		if errors.Is(err, keyring.ErrNotFound) {
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

func updateSheet(ctx context.Context, app *wakalog.Application, sheet string, rowIndex int, sheetId int64) error {

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

	if len(projectOptions) == 0 {

		return errNoProjects

	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select projects to get weekly activity from").
				Options(
					projectOptions...,
				).
				Value(&selectedProjects).Validate(func(options []string) error {
				if len(options) == 0 {
					return fmt.Errorf("Select a project (using x or spacebar) to proceed.")
				}

				return nil
			}),
		),
	)

	err = form.RunWithContext(ctx)

	if err != nil {
		return fmt.Errorf("error generating project options: %w", err)
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
	var mostActiveDay int
	var mostActiveDuration time.Duration

	// Loop over period/days e.g. Mon-Fri
	for i, data := range summaries.Data {

		var totalTimeForDay time.Duration

		for _, project := range data.Projects {

			projectName := project.Name

			if slices.Contains(selectedProjects, projectName) {

				totalTime, _ := time.ParseDuration(fmt.Sprintf("%dh%dm%ds", project.Hours, project.Minutes, project.Seconds))

				totalTimeForDay += totalTime

			}

		}

		if i == 0 {
			mostActiveDay = i
			mostActiveDuration = totalTimeForDay
		} else {
			if totalTimeForDay > mostActiveDuration {
				mostActiveDay = i
				mostActiveDuration = totalTimeForDay
			}
		}

		totalTimePerDay = append(totalTimePerDay, totalTimeForDay)
	}

	var cummulativeTotalTime time.Duration

	for _, totalTime := range totalTimePerDay {

		// Skip days where user had no coding activity.
		// Daily Average computation according to WakaTime ignores days of no activity
		if totalTime <= time.Duration(0) {
			continue
		}

		daysWorked += 1
		cummulativeTotalTime += totalTime

	}

	dailyAverage := cummulativeTotalTime.Hours() / float64(daysWorked)

	data := []interface{}{time.Duration(dailyAverage * float64(time.Hour)).Round(time.Second).String(), startDate.AddDate(0, 0, mostActiveDay).Format("Mon 1 Jan"), cummulativeTotalTime.Round(time.Second).String()}
	valueRange.Values = append(valueRange.Values, data)
	valueRange.Range = writeRange

	valuesRequest.Data = append(valuesRequest.Data, &valueRange)

	_, err = app.Sheets.Spreadsheets.Values.BatchUpdate(wakasheets.SpreadsheetId, valuesRequest).Do()

	if err != nil {
		return fmt.Errorf("unable to write data on sheet: %w", err)
	}

	return nil
}
