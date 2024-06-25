package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/99designs/keyring"
	"github.com/Youngtard/wakalog/httpclient"
	sheetsService "github.com/Youngtard/wakalog/sheets"
	"github.com/Youngtard/wakalog/wakatime"
	"github.com/icza/gox/timex"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var spreadsheetId = "1-edSRsFzUep-wtkvxoHKCnwltkrr2THzVK8n46RXGoo"

var sheetsApiScopes = []string{
	"https://www.googleapis.com/auth/spreadsheets.readonly",
	"https://www.googleapis.com/auth/spreadsheets",
}

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()

	srv, err := setUpSheetsService(ctx)

	if err != nil {
		log.Fatalf("Unable to retrieve Sheets service: %v", err)
	}

	var wakatimeAccessToken string

	wakatimeKeyring, _ := keyring.Open(keyring.Config{
		ServiceName: "wakalogs",
	})

	if item, err := wakatimeKeyring.Get("wakatime_access_token"); err != nil {

		if errors.Is(err, keyring.ErrKeyNotFound) {

			err := startWakatimeAuthorization(ctx, wakatimeAccessToken, wakatimeKeyring)

			if err != nil {
				log.Fatalf("Error authorizing wakatime: %v", err)

			}

		} else {
			log.Fatalf("Error retrieiving wakatime token: %v", err)
		}

		// TODO
		if len(item.Data) == 0 {
			err := startWakatimeAuthorization(ctx, wakatimeAccessToken, wakatimeKeyring)

			if err != nil {
				log.Fatalf("Error authorizing wakatime: %v", err)

			}
		}

	} else {

		err := wakatime.GetAccessToken(&wakatimeAccessToken, wakatimeKeyring)

		if err != nil {
			log.Fatalf("Error retrieving wakatime access token: %v", err)
		}
	}

	now := time.Now()

	year, week := now.ISOWeek()

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

	startDate := timex.WeekStart(year, week-relevantWeekOffset)
	endDate := startDate.AddDate(0, 0, 4)

	hc := httpclient.NewClient(nil).WithAuthToken(wakatimeAccessToken)

	wc := wakatime.NewClient(hc)

	summaries, err := wc.GetSummaries(ctx, startDate, endDate)

	if err != nil {
		log.Fatalf("Error getting summaries: %v", err)
	}

	engineersRange := "FEB!B:B"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, engineersRange).MajorDimension("COLUMNS").Do()

	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	name := "Femi Sotonwa"
	var rowIndex int

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		for _, row := range resp.Values {

			for i, v := range row {

				if v == name {

					rowIndex = i + 2

				}
			}
		}
	}

	writeRange := fmt.Sprintf("JUNE!C%d", rowIndex)

	valuesRequest := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "RAW",
	}

	var valueRange sheets.ValueRange

	data := []interface{}{summaries.DailyAverage.Text, "", summaries.CumulativeTotal.Text}
	valueRange.Values = append(valueRange.Values, data)
	valueRange.Range = writeRange

	valuesRequest.Data = append(valuesRequest.Data, &valueRange)

	_, err = srv.Spreadsheets.Values.BatchUpdate(spreadsheetId, valuesRequest).Do()

	if err != nil {
		log.Fatalf("Unable to write data on sheet: %v", err)
	}

}

func startWakatimeAuthorization(ctx context.Context, accessToken string, ring keyring.Keyring) error {
	err := wakatime.BeginAuthorization(ctx, &accessToken)

	if err != nil {
		return fmt.Errorf("error authorizing wakatime: %v", err)
	}

	err = wakatime.StoreAccessToken(accessToken, ring)

	if err != nil {

		return fmt.Errorf("error storing wakatime token: %v", err)

	}

	return nil
}

func setUpSheetsService(ctx context.Context) (*sheets.Service, error) {

	b, err := os.ReadFile("credentials.json")

	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)

	}

	config, err := google.ConfigFromJSON(b, sheetsApiScopes...)

	if err != nil {

		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	client, err := sheetsService.GetClient(ctx, config)

	if err != nil {
		return nil, fmt.Errorf("error getting sheets client: %v", err)
	}

	return sheets.NewService(ctx, option.WithHTTPClient(client))

}
