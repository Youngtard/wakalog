package wakatime

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
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

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(apiKeyPrompt).
				Placeholder("Enter API Key...").
				Value(&apiKey).
				Validate(func(str string) error {
					if len(str) == 0 {
						return fmt.Errorf("API Key is required to proceed.")
					}
					return nil
				}).WithTheme(huh.ThemeBase()),
		),
	)

	err := form.RunWithContext(ctx)

	if err != nil {
		return "", fmt.Errorf("error generating api key input: %w", err)

	}

	apiKey = strings.TrimSpace(apiKey)

	err = StoreAPIKey(apiKey)

	if err != nil {

		return "", fmt.Errorf("error storing wakatime api key: %w", err)

	}

	return apiKey, nil

}
