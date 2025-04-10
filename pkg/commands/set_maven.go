package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/abcxyz/pkg/cli"
	"github.com/yolocs/artifact-registry-cred-helper/pkg/maven"
)

type SetMavenCommand struct {
	baseCommand

	commonFlags       *CommonFlags
	mavenSettingsPath string
	repoIDsOverride   []string
}

func (c *SetMavenCommand) Desc() string {
	return "Set the credential in the Maven settings.xml file for the given repos."
}

func (c *SetMavenCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

Set the credential in the Maven settings.xml file for the given repos.
By default, we use repository ID in format: artifactregistry-[project_id]-[repo_name] to be used in pom.xml.

  # Example: Set the credential in the default path ~/.m2/settings.xml
  # The repo ID will be artifactregistry-my-project-my-repo
  artifact-registry-cred-helper set-maven --repo-urls us-maven.pkg.dev/my-project/repo1

  # Example: Override the default settings path.
  # The repo ID will be artifactregistry-my-repo
  artifact-registry-cred-helper set-maven --repo-urls us-maven.pkg.dev/my-project/repo1 --maven-settings /home/user/.m2/settings.xml

  # Example: Override the repo IDs.
  # The repo ID will be my-artifact-registry
  artifact-registry-cred-helper set-maven --repo-ids-override my-artifact-registry
`
}

func (c *SetMavenCommand) Flags() *cli.FlagSet {
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

func (c *SetMavenCommand) Run(ctx context.Context, args []string) (err error) {
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

	settings, err := maven.Open(c.mavenSettingsPath)
	if err != nil {
		return fmt.Errorf("failed to open Maven settings.xml file: %w", err)
	}

	// Immediately run once.
	if err := c.runOnce(ctx, settings); err != nil {
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
				if err := c.runOnce(ctx, settings); err != nil {
					return fmt.Errorf("failed to refresh credential: %w", err)
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}

func (c *SetMavenCommand) runOnce(ctx context.Context, settings authConfig) (err error) {
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
		k, err := c.getEncodedJSONKey(c.commonFlags.jsonKeyPath)
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

	token, err := c.getAuthToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}
	settings.SetToken(repoIDs, token)

	return nil
}
