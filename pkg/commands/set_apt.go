package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/abcxyz/pkg/cli"
	"github.com/yolocs/artifact-registry-cred-helper/pkg/apt"
	"github.com/yolocs/artifact-registry-cred-helper/pkg/auth"
)

type SetAptCommand struct {
	cli.BaseCommand

	commonFlags *CommonFlags
	configName  string
}

func (c *SetAptCommand) Desc() string {
	return "Set the credential in /etc/apt/auth.conf.d for the given repos."
}

func (c *SetAptCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

This command MUST be run in 'sudo -E' mode.

Set the credential in /etc/apt/auth.conf.d for the given repos.
All Artifact Registry credentials will be removed from the auth config before setting the new hosts.

  # Example: Set the credential in the default path /etc/apt/auth.conf.d/artifact-registry.conf
  artifact-registry-cred-helper set-apt --repo-urls us-apt.pkg.dev/my-project/repo1

  # Example: Override the default auth config path.
  artifact-registry-cred-helper set-apt --repo-urls us-apt.pkg.dev/my-project/repo1 --config-name my-repo.conf
`
}

func (c *SetAptCommand) Flags() *cli.FlagSet {
	c.commonFlags = &CommonFlags{}
	set := c.commonFlags.setSection(c.NewFlagSet())

	sec := set.NewSection("APT OPTIONS")
	sec.StringVar(&cli.StringVar{
		Name:    "config-name",
		Usage:   "The name of the config file under /etc/apt/auth.conf.d",
		Target:  &c.configName,
		EnvVar:  "AR_CRED_HELPER_APT_AUTH_CONFIG",
		Default: "artifact-registry.conf",
	})

	return set
}

func (c *SetAptCommand) Run(ctx context.Context, args []string) (err error) {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	if err := c.commonFlags.validate(); err != nil {
		return err
	}

	// Immediately run once.
	if err := c.runOnce(ctx); err != nil {
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
				if err := c.runOnce(ctx); err != nil {
					return fmt.Errorf("failed to refresh credential: %w", err)
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}

func (c *SetAptCommand) runOnce(ctx context.Context) (err error) {
	cfg, err := apt.Open(c.configName)
	if err != nil {
		return fmt.Errorf("failed to open apt auth config file: %w", err)
	}
	defer func() {
		if closeErr := cfg.Close(); err == nil {
			err = closeErr
		}
	}()

	hosts, err := c.commonFlags.repoHosts()
	if err != nil {
		// No error is possible here because we have validated the flag.
		return err
	}

	if c.commonFlags.jsonKeyPath != "" {
		k, err := auth.EncodeJSONKey(c.commonFlags.jsonKeyPath)
		if err != nil {
			return fmt.Errorf("failed to encode JSON key: %w", err)
		}
		cfg.SetJSONKey(hosts, k)
		return nil
	}

	if c.commonFlags.accessTokenFromEnv != "" {
		token := os.Getenv(c.commonFlags.accessTokenFromEnv)
		if token == "" {
			return fmt.Errorf("failed to get access token from env var %q", c.commonFlags.accessTokenFromEnv)
		}
		cfg.SetToken(hosts, token)
		return nil
	}

	token, err := auth.Token(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}
	cfg.SetToken(hosts, token)

	return nil
}
