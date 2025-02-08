package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/abcxyz/pkg/cli"
	"github.com/yolocs/artifact-registry-cred-helper/pkg/auth"
)

type authConfig interface {
	SetToken([]string, string)
	SetJSONKey([]string, string)
	Close() error
}

type authTokenGetter func(context.Context) (string, error)
type encodedJSONKeyGetter func(string) (string, error)

var (
	defaultAuthTokenGetter      = auth.Token
	defaultEncodedJSONKeyGetter = auth.EncodeJSONKey
)

type baseCommand struct {
	cli.BaseCommand

	getAuthToken      authTokenGetter
	getEncodedJSONKey encodedJSONKeyGetter
}

var rootCmd = func() cli.Command {
	return &cli.RootCommand{
		Name:    "artifact-registry-cred-helper",
		Version: "dev",
		Commands: map[string]cli.CommandFactory{
			"get": func() cli.Command {
				return &GetCommand{baseCommand: baseCommand{getAuthToken: defaultAuthTokenGetter, getEncodedJSONKey: defaultEncodedJSONKeyGetter}}
			},
			"set-netrc": func() cli.Command {
				return &SetNetRCCommand{baseCommand: baseCommand{getAuthToken: defaultAuthTokenGetter, getEncodedJSONKey: defaultEncodedJSONKeyGetter}}
			},
			"set-maven": func() cli.Command {
				return &SetMavenCommand{baseCommand: baseCommand{getAuthToken: defaultAuthTokenGetter, getEncodedJSONKey: defaultEncodedJSONKeyGetter}}
			},
			"set-apt": func() cli.Command {
				return &SetAptCommand{baseCommand: baseCommand{getAuthToken: defaultAuthTokenGetter, getEncodedJSONKey: defaultEncodedJSONKeyGetter}}
			},
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
