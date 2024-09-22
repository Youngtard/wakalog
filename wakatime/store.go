package wakatime

import (
	"fmt"

	"github.com/99designs/keyring"
	"github.com/Youngtard/wakalog/pkg/store"
)

func StoreAPIKey(apiKey string, ring keyring.Keyring) error {

	item := keyring.Item{Key: "wakatime_api_key", Data: []byte(apiKey)}

	err := ring.Set(item)

	if err != nil {
		return err
	}

	return nil

}

func GetAPIKey(apiKeyDest *string) error {

	ring, err := store.Keyring()

	if err != nil {
		return err
	}

	item, err := ring.Get("wakatime_api_key")

	if err != nil {
		return fmt.Errorf("error retreiving wakatime api key: %w", err)
	}

	*apiKeyDest = string(item.Data)

	return nil

}
