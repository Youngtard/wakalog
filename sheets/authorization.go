package sheets

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/int128/oauth2cli"
	"github.com/int128/oauth2cli/oauth2params"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/sync/errgroup"
)

var GoogleCredentials embed.FS

var scopes = []string{
	"https://www.googleapis.com/auth/spreadsheets.readonly",
	"https://www.googleapis.com/auth/spreadsheets",
}

func GetClient(ctx context.Context) (*http.Client, error) {

	config, err := getConfig()

	if err != nil {
		return nil, fmt.Errorf("error getting google config %w", err)
	}

	var token *oauth2.Token

	token, err = retrieveTokenFromFile()

	if err != nil {
		token, err = beginAuthorization(ctx)

		if err != nil {

			return nil, fmt.Errorf("error authorizing with sheets api: %v", err)

		}

	}

	// TODO no need to reauthorize after n days because refresh token doesn't expire?
	if time.Now().After(token.Expiry.AddDate(0, 0, 6)) {
		token, err = beginAuthorization(ctx)

		if err != nil {

			return nil, fmt.Errorf("error authorizing with sheets api: %v", err)

		}

	}

	// TODO no need to save if token already exists?
	saveToken(token)

	return config.Client(ctx, token), nil

}

func getConfig() (*oauth2.Config, error) {
	credentials, err := GoogleCredentials.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(credentials, scopes...)

	if err != nil {

		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	return config, nil
}

// TODO context timeout
func beginAuthorization(context context.Context) (*oauth2.Token, error) {

	config, err := getConfig()

	if err != nil {
		return nil, fmt.Errorf("error getting configuration: %w", err)
	}

	var token *oauth2.Token

	clientID := config.ClientID
	clientSecret := config.ClientSecret

	authURL := config.Endpoint.AuthURL
	tokenURL := config.Endpoint.TokenURL

	if clientID == "" {
		return nil, fmt.Errorf("sheets Client ID is required")

	}

	if clientSecret == "" {
		return nil, fmt.Errorf("sheet Client Secret is required")

	}

	pkce, err := oauth2params.NewPKCE()

	if err != nil {
		return nil, fmt.Errorf("error: %v", err)
	}

	ready := make(chan string, 1)
	defer close(ready)

	// Generate nonce to use as state parameter https://auth0.com/docs/secure/attack-protection/state-parameters
	nonceBytes := make([]byte, 64)
	_, err = io.ReadFull(rand.Reader, nonceBytes)
	if err != nil {
		return nil, fmt.Errorf("error generating random state parameter: %v", err)
	}
	randomStateValue := base64.URLEncoding.EncodeToString(nonceBytes)

	cfg := oauth2cli.Config{
		OAuth2Config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
			Scopes: scopes,
		},
		AuthCodeOptions:        pkce.AuthCodeOptions(),
		TokenRequestOptions:    pkce.TokenRequestOptions(),
		RedirectURLHostname:    "localhost",
		LocalServerBindAddress: []string{"127.0.0.1:8080"},
		LocalServerReadyChan:   ready,
		State:                  randomStateValue,
	}

	eg, ctx := errgroup.WithContext(context)

	eg.Go(func() error {
		select {
		case url := <-ready:
			if err := browser.OpenURL(url); err != nil {

				return fmt.Errorf("could not open the browser: %v", err)
			}
			return nil
		case <-ctx.Done():
			return fmt.Errorf("context done while waiting for authorization: %w", ctx.Err())
		}
	})

	eg.Go(func() error {
		token, err = oauth2cli.GetToken(ctx, cfg)
		if err != nil {
			return fmt.Errorf("could not get a token: %w", err)
		}

		return nil
	})

	if err := eg.Wait(); err != nil {

		return nil, fmt.Errorf("authorization error: %v", err)
	}

	if token != nil {
		return token, nil
	} else {
		return nil, fmt.Errorf("error authorizing sheets")
	}

}
