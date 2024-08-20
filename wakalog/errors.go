package wakalog

import "errors"

var ErrGeneric = errors.New("an error occurred")

// ErrWakaTimeTokenNotFound is returned when WakaTime token is not found in storage (Keyring)
var ErrWakaTimeTokenNotFound = errors.New("wakatime token not found")

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
