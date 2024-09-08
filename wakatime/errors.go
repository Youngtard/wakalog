package wakatime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Youngtard/wakalog/httpclient"
)

type WakaTimeError struct {
	StatusCode   int      `json:"-"`
	ErrorMessage *string  `json:"error"`
	Errors       []string `json:"errors"`
}

func (we *WakaTimeError) Error() string {

	switch {
	case we.StatusCode == http.StatusTooManyRequests:
		return "rate Limit reached"

	case we.StatusCode == http.StatusPaymentRequired:
		return "your plan doesn't cover this feature"

	default:
		if we.ErrorMessage != nil {
			return *we.ErrorMessage
		} else if we.Errors != nil {
			return strings.Join(we.Errors, ",")
		}

		return "an error occurred"
	}

}

func handleWakaTimeError(serverError *httpclient.ServerError) error {
	body, err := io.ReadAll(serverError.Body)

	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	wakatimeError := &WakaTimeError{StatusCode: serverError.StatusCode}

	if body != nil {
		err = json.Unmarshal(body, wakatimeError)

		if err != nil {
			return fmt.Errorf("error unmarshalling response body: %w", err)
		}
	}

	return wakatimeError

}
