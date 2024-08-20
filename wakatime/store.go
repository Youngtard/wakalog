package wakatime

import (
	"fmt"

	"github.com/99designs/keyring"
	"github.com/Youngtard/wakalog/pkg/store"
)

func StoreAccessToken(token string, ring keyring.Keyring) error {

	item := keyring.Item{Key: "wakatime_access_token", Data: []byte(token)}

	err := ring.Set(item)

	if err != nil {
		return err
	}

	return nil

}

func GetAccessToken(tokenDest *string) error {

	ring, err := store.Keyring()

	if err != nil {
		return err
	}

	item, err := ring.Get("wakatime_access_token")

	if err != nil {
		return fmt.Errorf("error retreiving wakatime access token: %w", err)
	}

	*tokenDest = string(item.Data)

	return nil

}
