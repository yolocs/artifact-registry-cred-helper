package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/abcxyz/pkg/cli"
	"github.com/yolocs/artifact-registry-cred-helper/pkg/npmrc"
)

type SetNPMCommand struct {
	baseCommand

	commonFlags *CommonFlags
	npmrcPath   string
	scope       string
}

func (c *SetNPMCommand) Desc() string {
	return "Set the credential in the .npmrc file for the given repos."
}

func (c *SetNPMCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

Set the credential in the .npmrc file for the given repos.

  # Example: Set the credential in the default path ~/.npmrc
  artifact-registry-cred-helper set-npmrc --repo-urls us-go.pkg.dev/my-project/repo1

  # Example: Override the default .npmrc path
  artifact-registry-cred-helper set-npmrc --repo-urls us-go.pkg.dev/my-project/repo1 --netrc /home/user/.netrc

  # Example: Set the scope for the given repos
  artifact-registry-cred-helper set-npmrc --repo-urls us-go.pkg.dev/my-project/repo1 --scope @my-scope
`
}

func (c *SetNPMCommand) Flags() *cli.FlagSet {
	c.commonFlags = &CommonFlags{}
	set := c.commonFlags.setSection(c.NewFlagSet())

	sec := set.NewSection("NPMRC OPTIONS")
	sec.StringVar(&cli.StringVar{
		Name:   "npmrc",
		Usage:  "The path to the .npmrc file. Default to the system default path.",
		Target: &c.npmrcPath,
		EnvVar: "AR_CRED_HELPER_NPMRC",
	})
	sec.StringVar(&cli.StringVar{
		Name:   "scope",
		Usage:  "The scope for the given repos.",
		Target: &c.scope,
		EnvVar: "AR_CRED_HELPER_SCOPE",
	})

	return set
}

func (c *SetNPMCommand) Run(ctx context.Context, args []string) (err error) {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	if err := c.commonFlags.validate(); err != nil {
		return err
	}

	nrc, err := npmrc.Open(c.npmrcPath, c.scope)
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

func (c *SetNPMCommand) runOnce(ctx context.Context, config authConfig) (err error) {
	defer func() {
		if closeErr := config.Close(); err == nil {
			err = closeErr
		}
	}()

	if c.commonFlags.jsonKeyPath != "" {
		k, err := c.getEncodedJSONKey(c.commonFlags.jsonKeyPath)
		if err != nil {
			return fmt.Errorf("failed to encode JSON key: %w", err)
		}
		config.SetJSONKey(c.commonFlags.repoURLs, k)
		return nil
	}

	if c.commonFlags.accessTokenFromEnv != "" {
		token := os.Getenv(c.commonFlags.accessTokenFromEnv)
		if token == "" {
			return fmt.Errorf("failed to get access token from env var %q", c.commonFlags.accessTokenFromEnv)
		}
		config.SetToken(c.commonFlags.repoURLs, token)
		return nil
	}

	token, err := c.getAuthToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}
	config.SetToken(c.commonFlags.repoURLs, token)

	return nil
}
