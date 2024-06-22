package wakatime

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/int128/oauth2cli"
	"github.com/int128/oauth2cli/oauth2params"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

var authURL = "https://waksatime.com/oauth/authorize"
var tokenURL = "https://wakatime.com/oauth/token"

func BeginAuthorization(ctx context.Context, tokenDest *string) error {

	appID := os.Getenv("WAKATIME_APP_ID")
	appSecret := os.Getenv("WAKATIME_APP_SECRET")

	if appID == "" {
		return fmt.Errorf("wakatime App ID is required")

	}

	if appSecret == "" {
		return fmt.Errorf("wakatime App Secret is required")

	}

	pkce, err := oauth2params.NewPKCE()

	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	ready := make(chan string, 1)
	defer close(ready)

	// Generate nonce to use as state parameter https://auth0.com/docs/secure/attack-protection/state-parameters
	nonceBytes := make([]byte, 64)
	_, err = io.ReadFull(rand.Reader, nonceBytes)
	if err != nil {
		return fmt.Errorf("error generating random state parameter: %v", err)
	}
	randomStateValue := base64.URLEncoding.EncodeToString(nonceBytes)

	// TODO make reusable

	cfg := oauth2cli.Config{
		OAuth2Config: oauth2.Config{
			ClientID:     appID,
			ClientSecret: appSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
			Scopes: []string{"read_summaries"},
		},
		AuthCodeOptions:        pkce.AuthCodeOptions(),
		TokenRequestOptions:    pkce.TokenRequestOptions(),
		RedirectURLHostname:    "localhost",
		LocalServerBindAddress: []string{"127.0.0.1:8080"},
		LocalServerReadyChan:   ready,
		State:                  randomStateValue,
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*60)

	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)
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
		token, err := oauth2cli.GetToken(ctx, cfg)
		if err != nil {
			return fmt.Errorf("could not get a token: %w", err)
		}

		*tokenDest = token.AccessToken
		return nil
	})

	if err := eg.Wait(); err != nil {

		return fmt.Errorf("authorization error: %v", err)
	}

	return nil

}

// func BeginAuthorization(ctx context.Context) (string, error) {

// 	var accessToken string

// 	appID := os.Getenv("WAKATIME_APP_ID")
// 	appSecret := os.Getenv("WAKATIME_APP_SECRET")

// 	if appID == "" {
// 		return "", fmt.Errorf("wakatime App ID is required")

// 	}

// 	if appSecret == "" {
// 		return "", fmt.Errorf("wakatime App Secret is required")

// 	}

// 	pkce, err := oauth2params.NewPKCE()

// 	if err != nil {
// 		return "", fmt.Errorf("error: %v", err)
// 	}

// 	ready := make(chan string, 1)
// 	defer close(ready)

// 	// Generate nonce to use as state parameter https://auth0.com/docs/secure/attack-protection/state-parameters
// 	nonceBytes := make([]byte, 64)
// 	_, err = io.ReadFull(rand.Reader, nonceBytes)
// 	if err != nil {
// 		return "", fmt.Errorf("error generating random state parameter: %v", err)
// 	}
// 	randomStateValue := base64.URLEncoding.EncodeToString(nonceBytes)

// 	// TODO make reusable

// 	cfg := oauth2cli.Config{
// 		OAuth2Config: oauth2.Config{
// 			ClientID:     appID,
// 			ClientSecret: appSecret,
// 			Endpoint: oauth2.Endpoint{
// 				AuthURL:  authURL,
// 				TokenURL: tokenURL,
// 			},
// 			Scopes: []string{"read_summaries"},
// 		},
// 		AuthCodeOptions:        pkce.AuthCodeOptions(),
// 		TokenRequestOptions:    pkce.TokenRequestOptions(),
// 		RedirectURLHostname:    "localhost",
// 		LocalServerBindAddress: []string{"127.0.0.1:8080"},
// 		LocalServerReadyChan:   ready,
// 		State:                  randomStateValue,
// 	}

// 	eg, ctx := errgroup.WithContext(ctx)
// 	eg.Go(func() error {
// 		select {
// 		case url := <-ready:
// 			if err := browser.OpenURL(url); err != nil {

// 				return fmt.Errorf("could not open the browser: %v", err)
// 			}
// 			return nil
// 		case <-ctx.Done():
// 			return fmt.Errorf("context done while waiting for authorization: %w", ctx.Err())
// 		}
// 	})
// 	eg.Go(func() error {
// 		token, err := oauth2cli.GetToken(ctx, cfg)
// 		if err != nil {
// 			return fmt.Errorf("could not get a token: %w", err)
// 		}

// 		accessToken = token.AccessToken
// 		return nil
// 	})
// 	if err := eg.Wait(); err != nil {

// 		return "", fmt.Errorf("authorization error: %v", err)
// 	}

// 	if accessToken != "" {
// 		return accessToken, nil
// 	} else {
// 		return "", fmt.Errorf("error authorizing wakatime")
// 	}

// }
