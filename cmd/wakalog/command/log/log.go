package log

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"github.com/Youngtard/wakalog/pkg/cmdutil"
	"github.com/Youngtard/wakalog/sheets"
	"github.com/Youngtard/wakalog/wakalog"
	"github.com/Youngtard/wakalog/wakatime"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

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
						sheetsToken, err = sheets.Authorize(ctx)

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
				sheetsToken, err = sheets.Authorize(ctx)

				if err != nil {

					return &wakalog.AuthError{Err: fmt.Errorf("error authenticating with Google Sheets")}

				}

			}

			app.InitializeWakaTime(wakatimeToken)
			app.InitializeSheets(ctx, sheetsToken)

			return nil

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Log something")

			return nil

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
	token, err := sheets.RetrieveTokenFromFile()

	if err != nil {
		return nil, wakalog.ErrSheetsTokenNotFound
	}

	return token, nil
}
