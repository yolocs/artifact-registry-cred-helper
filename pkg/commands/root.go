package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/abcxyz/pkg/cli"
)

var rootCmd = func() cli.Command {
	return &cli.RootCommand{
		Name:    "artifact-registry-cred-helper",
		Version: "dev",
		Commands: map[string]cli.CommandFactory{
			"get":       func() cli.Command { return &GetCommand{} },
			"set-netrc": func() cli.Command { return &SetNetRCCommand{} },
			"set-maven": func() cli.Command { return &SetMavenSettings{} },
		},
	}
}

// Run executes the CLI.
func Run(ctx context.Context, args []string) error {
	return rootCmd().Run(ctx, args) //nolint:wrapcheck // Want passthrough
}

func validateHosts(hosts []string) error {
	if len(hosts) <= 0 {
		return errors.New("no host specified")
	}

	var merr error
	for _, h := range hosts {
		if !strings.HasSuffix(h, ".pkg.dev") {
			merr = errors.Join(merr, fmt.Errorf("host %q doesn't have domain '.pkg.dev'", h))
		}
	}
	return merr
}
