package wakatime

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

var serviceName string = "wakalog"
var userName string = "wakatime_api_key"

func StoreAPIKey(apiKey string) error {

	err := keyring.Set(serviceName, userName, apiKey)

	if err != nil {
		return err
	}

	return nil

}

func GetAPIKey(apiKeyDest *string) error {

	apiKey, err := keyring.Get(serviceName, userName)

	if err != nil {
		return fmt.Errorf("error retreiving wakatime api key: %w", err)
	}

	*apiKeyDest = apiKey

	return nil

}
