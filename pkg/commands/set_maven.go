package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/abcxyz/pkg/cli"
	"github.com/yolocs/artifact-registry-cred-helper/pkg/auth"
	"github.com/yolocs/artifact-registry-cred-helper/pkg/maven"
)

type SetMavenSettings struct {
	cli.BaseCommand

	commonFlags       *CommonFlags
	mavenSettingsPath string
	repoIDsOverride   []string
}

func (c *SetMavenSettings) Desc() string {
	return "Set the credential in the .netrc file for the given host."
}

func (c *SetMavenSettings) Help() string {
	return `
Usage: {{ COMMAND }} [options]

Set the credential in the Maven settings.xml file for the given repos.

TODO: better docs
`
}

func (c *SetMavenSettings) Flags() *cli.FlagSet {
	c.commonFlags = &CommonFlags{}
	set := c.commonFlags.setSection(c.NewFlagSet())

	sec := set.NewSection("MAVEN OPTIONS")
	sec.StringVar(&cli.StringVar{
		Name:   "maven-settings",
		Usage:  "The path to the Maven settings.xml file. Default to ~/.m2/settings.xml.",
		Target: &c.mavenSettingsPath,
		EnvVar: "AR_CRED_HELPER_MAVEN_SETTINGS",
	})
	sec.StringSliceVar(&cli.StringSliceVar{
		Name:    "repo-ids-override",
		Usage:   "Override the repo IDs that are used in pom.xml.",
		Target:  &c.repoIDsOverride,
		EnvVar:  "AR_CRED_HELPER_MAVEN_REPO_IDS_OVERRIDE",
		Example: "my-artifact-registry",
	})

	return set
}

func (c *SetMavenSettings) Run(ctx context.Context, args []string) (err error) {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if len(c.repoIDsOverride) > 0 {
		if err := c.commonFlags.validateWithoutURLs(); err != nil {
			return err
		}
	} else {
		if err := c.commonFlags.validate(); err != nil {
			return err
		}
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
			case <-ticker.C: // Run periodically
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

func (c *SetMavenSettings) runOnce(ctx context.Context) (err error) {
	settings, err := maven.Open(c.mavenSettingsPath)
	if err != nil {
		return fmt.Errorf("failed to open Maven settings.xml file: %w", err)
	}
	defer func() {
		if closeErr := settings.Close(); err == nil {
			err = closeErr
		}
	}()

	repoIDs := c.repoIDsOverride
	if len(repoIDs) <= 0 {
		for _, u := range c.commonFlags.parsedURLs {
			repoIDs = append(repoIDs, maven.DefaultRepoID(u))
		}
	}

	if c.commonFlags.jsonKeyPath != "" {
		k, err := auth.EncodeJSONKey(c.commonFlags.jsonKeyPath)
		if err != nil {
			return fmt.Errorf("failed to encode JSON key: %w", err)
		}
		settings.SetJSONKey(repoIDs, k)
		return nil
	}

	if c.commonFlags.accessTokenFromEnv != "" {
		token := os.Getenv(c.commonFlags.accessTokenFromEnv)
		if token == "" {
			return fmt.Errorf("failed to get access token from env var %q", c.commonFlags.accessTokenFromEnv)
		}
		settings.SetToken(repoIDs, token)
		return nil
	}

	token, err := auth.Token(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}
	settings.SetToken(repoIDs, token)

	return nil
}
