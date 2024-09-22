package wakalog

import "errors"

var ErrGeneric = errors.New("an error occurred")

// ErrWakaTimeAPIKeyNotFound is returned when WakaTime API Key is not found in storage (Keyring)
var ErrWakaTimeAPIKeyNotFound = errors.New("wakatime API key not found")

var ErrSheetsTokenNotFound = errors.New("sheets token not found")

type FlagError struct {
	Err error
}

func (fe *FlagError) Error() string {

	return fe.Err.Error()

}

type AuthError struct {
	Err error
}

func (ae *AuthError) Error() string {

	return ae.Err.Error()

}
