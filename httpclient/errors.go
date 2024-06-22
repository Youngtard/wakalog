package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type WakaTimeError struct {
	Response   *http.Response `json:"-"`
	StatusCode int            `json:"-"`
	// todo what if not in response
	Message string `json:"error"`
}

func (r *WakaTimeError) Error() string {
	return fmt.Sprintf("Error (%d): %s", r.StatusCode, r.Message)
}

type RateLimitError struct {
	Response *http.Response `json:"-"`
	Message  string         `json:"error"`
}

func (r *RateLimitError) Error() string {
	return "You are being rate limited, try making fewer than 10 requests per second on average over any 5 minute period."
}

func handleWakaTimeError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	wakatimeError := &WakaTimeError{Response: resp, StatusCode: resp.StatusCode}

	if body != nil {
		err = json.Unmarshal(body, wakatimeError)

		if err != nil {
			return err
		}
	}

	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return &RateLimitError{Response: wakatimeError.Response, Message: wakatimeError.Message}

	// case resp.StatusCode == http.StatusPaymentRequired:
	// 	return &RateLimitError{Response: wakatimeError.Response, Message: wakatimeError.Message}

	default:
		return wakatimeError
	}

}
