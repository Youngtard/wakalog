package wakatime

import (
	"context"
	"fmt"

	"github.com/Youngtard/wakalog/pkg/interact"
	"github.com/savioxavier/termlink"
)

func Authorize(ctx context.Context) (string, error) {

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

	err := interact.TextInput(apiKeyPrompt, &apiKey)

	if err != nil {
		return "", fmt.Errorf("error generating api key input")
	}

	err = StoreAPIKey(apiKey)

	if err != nil {

		return "", fmt.Errorf("error storing wakatime api key: %v", err)

	}

	return apiKey, nil

}
