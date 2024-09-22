package wakatime

import (
	"context"
	"fmt"

	"github.com/99designs/keyring"
	"github.com/Youngtard/wakalog/pkg/interact"
	"github.com/savioxavier/termlink"
)

func Authorize(ctx context.Context) (string, error) {

	wakatimeKeyring, err := keyring.Open(keyring.Config{
		ServiceName: "wakalog",
	})

	if err != nil {
		return "", err
	}

	var apiKey string

	apiKeyUrl := "https://wakatime.com/settings/api-key"

	apiKeyLink := termlink.ColorLink(apiKeyUrl, apiKeyUrl, "italic blue")
	prompt := "Enter your WakaTime API Key to proceed."
	var apiKeyPrompt string

	if termlink.SupportsHyperlinks() {
		apiKeyPrompt = fmt.Sprintf("%s %s", prompt, apiKeyLink)
	} else {
		apiKeyPrompt = prompt
	}

	err = interact.TextInput(apiKeyPrompt, &apiKey)

	if err != nil {
		return "", fmt.Errorf("error generating api key input")
	}

	err = StoreAPIKey(apiKey, wakatimeKeyring)

	if err != nil {

		return "", fmt.Errorf("error storing wakatime api key: %v", err)

	}

	return apiKey, nil

}
