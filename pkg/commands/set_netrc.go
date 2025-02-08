package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/abcxyz/pkg/cli"
	"github.com/yolocs/artifact-registry-cred-helper/pkg/netrc"
)

type SetNetRCCommand struct {
	baseCommand

	commonFlags *CommonFlags
	netrcPath   string
}

func (c *SetNetRCCommand) Desc() string {
	return "Set the credential in the .netrc file for the given repos."
}

func (c *SetNetRCCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

Set the credential in the .netrc file for the given host(s).
All Artifact Registry credentials will be removed from the .netrc file before setting the new hosts.

  # Example: Set the credential in the default path ~/.netrc
  artifact-registry-cred-helper set-netrc --repo-urls us-go.pkg.dev/my-project/repo1

  # Example: Override the default .netrc path
  artifact-registry-cred-helper set-netrc --repo-urls us-go.pkg.dev/my-project/repo1 --netrc /home/user/.netrc
`
}

func (c *SetNetRCCommand) Flags() *cli.FlagSet {
	c.commonFlags = &CommonFlags{}
	set := c.commonFlags.setSection(c.NewFlagSet())

	sec := set.NewSection("NETRC OPTIONS")
	sec.StringVar(&cli.StringVar{
		Name:   "netrc",
		Usage:  "The path to the .netrc file. Default to the system default path.",
		Target: &c.netrcPath,
		EnvVar: "AR_CRED_HELPER_NETRC",
	})

	return set
}

func (c *SetNetRCCommand) Run(ctx context.Context, args []string) (err error) {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	if err := c.commonFlags.validate(); err != nil {
		return err
	}

	nrc, err := netrc.Open(c.netrcPath)
	if err != nil {
		return fmt.Errorf("failed to open .netrc file: %w", err)
	}

	// Immediately run once.
	if err := c.runOnce(ctx, nrc); err != nil {
		return fmt.Errorf("failed to set credential: %w", err)
	}

	// Start background refresh if enabled.
	if c.commonFlags.backgroundRefreshInterval > 0 {
		ctx, cancel := context.WithTimeout(ctx, c.commonFlags.backgroundRefreshDuration)
		defer cancel()
		ticker := time.NewTicker(c.commonFlags.backgroundRefreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := c.runOnce(ctx, nrc); err != nil {
					return fmt.Errorf("failed to refresh credential: %w", err)
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}

func (c *SetNetRCCommand) runOnce(ctx context.Context, nrc authConfig) (err error) {
	defer func() {
		if closeErr := nrc.Close(); err == nil {
			err = closeErr
		}
	}()

	hosts, err := c.commonFlags.repoHosts()
	if err != nil {
		// No error is possible here because we have validated the flag.
		return err
	}

	if c.commonFlags.jsonKeyPath != "" {
		k, err := c.getEncodedJSONKey(c.commonFlags.jsonKeyPath)
		if err != nil {
			return fmt.Errorf("failed to encode JSON key: %w", err)
		}
		nrc.SetJSONKey(hosts, k)
		return nil
	}

	if c.commonFlags.accessTokenFromEnv != "" {
		token := os.Getenv(c.commonFlags.accessTokenFromEnv)
		if token == "" {
			return fmt.Errorf("failed to get access token from env var %q", c.commonFlags.accessTokenFromEnv)
		}
		nrc.SetToken(hosts, token)
		return nil
	}

	token, err := c.getAuthToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}
	nrc.SetToken(hosts, token)

	return nil
}
