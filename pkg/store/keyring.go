package store

import (
	"fmt"

	"github.com/99designs/keyring"
)

func Keyring() (keyring.Keyring, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: "wakalog",
	})

	if err != nil {
		return nil, fmt.Errorf("error opening keyring: %w", err)
	}

	return ring, nil
}
