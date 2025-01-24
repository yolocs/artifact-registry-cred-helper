package commands

import (
	"context"
	"fmt"

	"github.com/abcxyz/pkg/cli"
	"github.com/yolocs/artifact-registry-cred-helper/pkg/auth"
	"github.com/yolocs/artifact-registry-cred-helper/pkg/netrc"
)

type SetNetRCCommand struct {
	cli.BaseCommand

	hosts       []string
	jsonKey     string
	netrcPath   string
	refreshOnly bool
	appendOnly  bool
}

func (c *SetNetRCCommand) Desc() string {
	return "Set the credential in the .netrc file for the given host."
}

func (c *SetNetRCCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

Set the credential in the .netrc file for the given host(s).

By default, all Artifact Registry credentials will be removed from the .netrc file before adding the new hosts.

  # Existing Artifact Registry hosts will be removed.
  # The only one remaining will be us-go.pkg.dev.
  # If us-go.pkg.dev already exists, its credential will be refreshed.
  artifact-registry-cred-helper set-netrc --hosts=us-go.pkg.dev

To only refresh the existing hosts credentials, use --refresh flag.

  # Existing Artifact Registry hosts will be refreshed with new credentials.
  artifact-registry-cred-helper set-netrc --refresh

To append new hosts without touch existing hosts, use --append flag.

  # New host us-go.pkg.dev will be appended. Potentially causing a duplicate entry if us-go.pkg.dev already exists.
  artifact-registry-cred-helper set-netrc --hosts=us-go.pkg.dev --append
`
}

func (c *SetNetRCCommand) Flags() *cli.FlagSet {
	set := c.NewFlagSet()
	sec := set.NewSection("OPTIONS")

	sec.StringSliceVar(&cli.StringSliceVar{
		Name:    "hosts",
		Usage:   "The hosts to set the credential for. Must have domain '*.pkg.dev'.",
		Target:  &c.hosts,
		EnvVar:  "AR_CRED_HELPER_HOSTS",
		Example: "us-go.pkg.dev,eu-python.pkg.dev,asia-maven.pkg.dev",
	})

	sec.StringVar(&cli.StringVar{
		Name:   "json-key",
		Usage:  "The path to the JSON key of a service account used for authentication. Leave empty to use oauth token instead.",
		Target: &c.jsonKey,
		EnvVar: "AR_CRED_HELPER_JSON_KEY",
	})

	sec.StringVar(&cli.StringVar{
		Name:   "netrc",
		Usage:  "The path to the .netrc file. Default to the system default path.",
		Target: &c.netrcPath,
		EnvVar: "AR_CRED_HELPER_NETRC",
	})

	sec.BoolVar(&cli.BoolVar{
		Name:    "refresh",
		Usage:   "Specify this flag to refresh .netrc. Does not work with --json-key.",
		Target:  &c.refreshOnly,
		EnvVar:  "AR_CRED_HELPER_REFRESH_ONLY",
		Default: false,
	})

	sec.BoolVar(&cli.BoolVar{
		Name:    "append",
		Usage:   "Specify this flag to append to .netrc instead of removing Artifact Registry entries.",
		Target:  &c.appendOnly,
		EnvVar:  "AR_CRED_HELPER_APPEND_ONLY",
		Default: false,
	})

	return set
}

func (c *SetNetRCCommand) Run(ctx context.Context, args []string) error {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	nrc, err := netrc.Open(c.netrcPath)
	if err != nil {
		return fmt.Errorf("failed to open .netrc file: %w", err)
	}

	if c.refreshOnly {
		token, err := auth.Token(ctx)
		if err != nil {
			return fmt.Errorf("failed to get access token: %w", err)
		}
		nrc.Refresh(token)
		return nrc.Close()
	}

	if err := validateHosts(c.hosts); err != nil {
		return err
	}

	if c.jsonKey == "" {
		token, err := auth.Token(ctx)
		if err != nil {
			return fmt.Errorf("failed to get access token: %w", err)
		}
		nrc.SetToken(c.hosts, token, c.appendOnly)
	} else {
		k, err := auth.EncodeJSONKey(c.jsonKey)
		if err != nil {
			return fmt.Errorf("failed to encode JSON key: %w", err)
		}
		nrc.SetJSONKey(c.hosts, k, c.appendOnly)
	}

	return nrc.Close()
}
