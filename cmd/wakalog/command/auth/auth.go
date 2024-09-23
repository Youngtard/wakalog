package auth

import (
	"fmt"

	"github.com/Youngtard/wakalog/wakalog"
	"github.com/Youngtard/wakalog/wakatime"
	"github.com/spf13/cobra"
)

func NewAuthCmd(app *wakalog.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authorize WakaTime.",
		Long:  "Authorize WakaTime with API Key.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			_, err := wakatime.Authorize(cmd.Context())

			if err != nil {
				return &wakalog.AuthError{Err: fmt.Errorf("error authenticating with WakaTime: %w", err)}
			}

			fmt.Println("WakaTime auth successful!")

			return nil
		},
	}

	return cmd

}
