package wakalog

import (
	"context"
	"fmt"
	"net/http"

	"encoding/base64"

	"github.com/Youngtard/wakalog/httpclient"
	"github.com/Youngtard/wakalog/wakatime"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Application struct {
	WakaTime *wakatime.Client
	// TODO have a wrapper? conflicting with project sheets package
	Sheets *sheets.Service
}

func NewApplication(context context.Context) *Application {

	app := &Application{}

	return app

}

func (app *Application) InitializeWakaTime(apiKey string) {

	// TODO check if apiKey is not empty
	// TODO nil checks?

	encodedKey := base64.StdEncoding.EncodeToString([]byte(apiKey))

	hc := httpclient.NewClient(nil).WithBasicAuth(encodedKey)

	wc := wakatime.NewClient(hc)

	app.WakaTime = wc

}

func (app *Application) InitializeSheets(context context.Context, client *http.Client) error {

	srv, err := sheets.NewService(context, option.WithHTTPClient(client))

	if err != nil {
		return fmt.Errorf("error setting up sheets service: %w", err)
	}

	app.Sheets = srv

	return nil

}
