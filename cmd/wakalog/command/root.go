package command

import (
	"github.com/Youngtard/wakalog/wakalog"
	"github.com/spf13/cobra"
)

func NewRootCommand(app *wakalog.Application) *cobra.Command {
	cobra.OnInitialize()

	cmd := &cobra.Command{
		Use:           "wakalog <command> <subcommand> [flags]",
		Short:         "Log your WakaTime summaries",
		Long:          "A CLI to log your weekly WakaTime summaries to a Google Spreadsheet",
		SilenceUsage:  true,
		SilenceErrors: true,

		PersistentPreRun: func(cmd *cobra.Command, args []string) {

		},
		// Version:               fmt.Sprintf("%s, build %s", version.Version, version.GitCommit),

	}

	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {

		return &wakalog.FlagError{Err: err}

	})

	addCommands(cmd, app)

	return cmd

}
