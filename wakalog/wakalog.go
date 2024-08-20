package wakalog

import (
	"context"
	"fmt"

	"github.com/Youngtard/wakalog/httpclient"
	sheetsService "github.com/Youngtard/wakalog/sheets"
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

func NewApplication(context context.Context, wakaTimeToken string, sheetsToken *oauth2.Token) (*Application, error) {

	app := &Application{}

	app.InitializeWakaTime(wakaTimeToken)
	app.InitializeSheets(context, sheetsToken)
	return app, nil

}

func (app *Application) InitializeWakaTime(token string) {

	hc := httpclient.NewClient(nil).WithAuthToken(token)

	wc := wakatime.NewClient(hc)

	app.WakaTime = wc

}

func (app *Application) InitializeSheets(context context.Context, token *oauth2.Token) error {

	config, err := sheetsService.GetConfig()

	if err != nil {
		return fmt.Errorf("error getting confifiguration: %w", err)
	}

	client := config.Client(context, token)

	srv, err := sheets.NewService(context, option.WithHTTPClient(client))

	if err != nil {
		return fmt.Errorf("error setting up sheets service: %w", err)
	}

	app.Sheets = srv

	return nil

}
