package wakalog

import (
	"context"
	"fmt"

	"encoding/base64"

	"github.com/Youngtard/wakalog/httpclient"
	wakasheets "github.com/Youngtard/wakalog/sheets"
	"github.com/Youngtard/wakalog/wakatime"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Application struct {
	WakaTime *wakatime.Client
	// TODO have a wrapper? conflicting with project sheets package
	Sheets *sheets.Service
}

func NewApplication(context context.Context, wakaTimeAPIKey string, sheetsToken *oauth2.Token) (*Application, error) {

	app := &Application{}

	app.InitializeWakaTime(wakaTimeAPIKey)
	err := app.InitializeSheets(context, sheetsToken)

	if err != nil {
		return nil, fmt.Errorf("error initializing sheets service %w", err)
	}
	return app, nil

}

func (app *Application) InitializeWakaTime(apiKey string) {

	// TODO check if apiKey is not empty
	// TODO nil checks?

	encodedKey := base64.StdEncoding.EncodeToString([]byte(apiKey))

	hc := httpclient.NewClient(nil).WithBasicAuth(encodedKey)

	wc := wakatime.NewClient(hc)

	app.WakaTime = wc

}

func (app *Application) InitializeSheets(context context.Context, token *oauth2.Token) error {

	config, err := wakasheets.GetConfig()

	if err != nil {
		return fmt.Errorf("error getting google config %w", err)
	}

	client, err := wakasheets.GetClient(context, config)

	if err != nil {
		return fmt.Errorf("error getting google client %w", err)
	}

	srv, err := sheets.NewService(context, option.WithHTTPClient(client))

	if err != nil {
		return fmt.Errorf("error setting up sheets service: %w", err)
	}

	app.Sheets = srv

	return nil

}
