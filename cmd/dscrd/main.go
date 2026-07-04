// Command dscrd is an agent-facing Discord CLI.
package main

import (
	"errors"
	"fmt"
	"os"

	dscrd "github.com/howar31/dscrd"
	"github.com/howar31/dscrd/internal/api"
	"github.com/howar31/dscrd/internal/auth"
	"github.com/howar31/dscrd/internal/commands"
)

func main() {
	root := commands.NewRootCommand(dscrd.Version)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "dscrd:", err)
		var apiErr *api.APIError
		if errors.As(err, &apiErr) {
			os.Exit(apiErr.ExitCode())
		}
		var authErr *auth.AuthError
		if errors.As(err, &authErr) {
			os.Exit(authErr.ExitCode())
		}
		os.Exit(1)
	}
}
