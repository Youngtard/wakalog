package cmdutil

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// https://github.com/docker/cli/blob/master/cli/command/utils.go
func PromptForConfirmation(ctx context.Context, message string) (bool, error) {
	if message == "" {
		message = "Are you sure you want to proceed?"
	}
	message += " [y/N] "

	_, _ = fmt.Fprint(os.Stdout, message)

	result := make(chan bool)

	go func() {
		var res bool
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			answer := strings.TrimSpace(scanner.Text())
			if strings.EqualFold(answer, "y") {
				res = true
			}
		}
		result <- res
	}()

	select {
	case <-ctx.Done():
		_, _ = fmt.Fprintln(os.Stdout, "")
		return false, fmt.Errorf("confirmation prompt terminated")
	case r := <-result:
		return r, nil
	}
}
