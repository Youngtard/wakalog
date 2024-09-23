package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Youngtard/wakalog/cmd/wakalog/command"
	wakasheets "github.com/Youngtard/wakalog/sheets"
	"github.com/Youngtard/wakalog/wakalog"
	"github.com/Youngtard/wakalog/wakatime"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

//go:embed credentials.json
var googleCredentials embed.FS

func startCli(ctx context.Context, app *wakalog.Application) (*cobra.Command, error) {
	return command.NewRootCommand(app).ExecuteContextC(ctx)
}

func main() {
	wakasheets.GoogleCredentials = googleCredentials
	ctx := context.Background()

	db, err := openDb()
	if err != nil {
		log.Fatal("Error opening db: %w", err)
	}

	app := wakalog.NewApplication(ctx, db)

	if cmd, err := startCli(ctx, app); err != nil {
		errorLog := log.New(os.Stderr, "", 0)

		errorCode := 1

		var flagError *wakalog.FlagError
		var authError *wakalog.AuthError
		var wakatimeError *wakatime.WakaTimeError

		if errors.As(err, &authError) {
			errorLog.Println(err)
		} else if errors.As(err, &flagError) || strings.HasPrefix(err.Error(), "unknown command ") {

			errorLog.Println(err)

			if !strings.HasSuffix(err.Error(), "\n") {
				fmt.Fprintln(os.Stdout)
			}

			errorLog.Println(cmd.UsageString())
			os.Exit(1)

		} else if errors.As(err, &wakatimeError) {
			errorCode = wakatimeError.StatusCode
			errorLog.Printf("WakaTime Error: %s (%d)\n", wakatimeError, errorCode)
			os.Exit(errorCode)

		} else {
			errorLog.Printf("An error occurred: %s (%d)\n", err, errorCode)
			os.Exit(errorCode)
		}

	}

}

func openDb() (*bolt.DB, error) {

	db, err := bolt.Open("wakalog.db", 0600, &bolt.Options{Timeout: 5 * time.Second})

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	return db, nil

}
