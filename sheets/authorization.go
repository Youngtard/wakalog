package sheets

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/int128/oauth2cli"
	"github.com/int128/oauth2cli/oauth2params"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

var scopes = []string{
	"https://www.googleapis.com/auth/spreadsheets.readonly",
	"https://www.googleapis.com/auth/spreadsheets",
}

func GetClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {

	var token *oauth2.Token

	tokenPath := "token.json"

	token, err := retrieveTokenFromFile(tokenPath)

	if err != nil {
		token, err = startAuthorization(ctx, config)

		if err != nil {

			return nil, fmt.Errorf("error authorizing with sheets api: %v", err)

		}

	}

	// TODO after 7 days / tokenExpiry + 7 days
	if time.Now().After(token.Expiry) {
		token, err = startAuthorization(ctx, config)

		if err != nil {

			return nil, fmt.Errorf("error authorizing with sheets api: %v", err)

		}

	}

	saveToken(tokenPath, token)

	return config.Client(context.Background(), token), nil

}

// TODO token from keystring?
func retrieveTokenFromFile(filePath string) (*oauth2.Token, error) {
	tokenFile, err := os.Open(filePath)

	if err != nil {
		return nil, err
	}

	defer tokenFile.Close()

	tok := &oauth2.Token{}

	err = json.NewDecoder(tokenFile).Decode(tok)

	return tok, err

}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}

	defer f.Close()

	json.NewEncoder(f).Encode(token)
}

// TODO context timeout
func startAuthorization(context context.Context, config *oauth2.Config) (*oauth2.Token, error) {

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
