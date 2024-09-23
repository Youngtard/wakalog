package command

import (
	"github.com/Youngtard/wakalog/cmd/wakalog/command/auth"
	"github.com/Youngtard/wakalog/cmd/wakalog/command/log"
	"github.com/Youngtard/wakalog/wakalog"
	"github.com/spf13/cobra"
)

func addCommands(cmd *cobra.Command, app *wakalog.Application) {

	cmd.AddCommand(log.NewLogCommand(app))
	cmd.AddCommand(auth.NewAuthCmd(app))

}
