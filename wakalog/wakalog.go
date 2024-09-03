package wakalog

import (
	"context"
	"fmt"

	"github.com/Youngtard/wakalog/httpclient"
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
	err := app.InitializeSheets(context, sheetsToken)

	if err != nil {
		return nil, fmt.Errorf("error initializing sheets service %w", err)
	}
	return app, nil

}

func (app *Application) InitializeWakaTime(token string) {

	// TODO check if token is not empty
	// TODO nil checks?
	hc := httpclient.NewClient(nil).WithAuthToken(token)

	wc := wakatime.NewClient(hc)

	app.WakaTime = wc

}

func (app *Application) InitializeSheets(context context.Context, token *oauth2.Token) error {

	credentialsOption := option.WithCredentialsFile("service_account.json")

	srv, err := sheets.NewService(context, credentialsOption)

	if err != nil {
		return fmt.Errorf("error setting up sheets service: %w", err)
	}

	app.Sheets = srv

	return nil

}
