package commands

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/abcxyz/pkg/cli"
)

type CommonFlags struct {
	hosts                     []string
	accessTokenFromEnv        string
	jsonKeyPath               string
	backgroundRefreshInterval time.Duration
	backgroundRefreshDuration time.Duration
}

func (f *CommonFlags) validate() error {
	var merr error

	if len(f.hosts) <= 0 {
		merr = errors.Join(merr, fmt.Errorf("no host specified"))
	}

	for _, h := range f.hosts {
		if !strings.HasSuffix(h, ".pkg.dev") {
			merr = errors.Join(merr, fmt.Errorf("host %q doesn't have domain '.pkg.dev'", h))
		}
	}

	if f.backgroundRefreshInterval < 2*time.Minute && f.backgroundRefreshInterval > 0 {
		merr = errors.Join(merr, fmt.Errorf("background refresh interval must be at least 2 minutes"))
	}

	if f.jsonKeyPath != "" && f.accessTokenFromEnv != "" {
		merr = errors.Join(merr, fmt.Errorf("only one of --json-key or --access-token-from-env can be set"))
	}

	return merr
}

func (f *CommonFlags) setSection(set *cli.FlagSet) *cli.FlagSet {
	sec := set.NewSection("COMMON OPTIONS")

	sec.StringSliceVar(&cli.StringSliceVar{
		Name:    "hosts",
		Usage:   "REQUIRED. The hosts to set the credential for. Must have domain '*.pkg.dev'.",
		Target:  &f.hosts,
		EnvVar:  "AR_CRED_HELPER_HOSTS",
		Example: "us-go.pkg.dev,eu-python.pkg.dev,asia-maven.pkg.dev",
	})

	sec.StringVar(&cli.StringVar{
		Name:    "access-token-from-env",
		Usage:   "The env var name to load the access token from. This is useful when it's another process that produces the access token.",
		Target:  &f.accessTokenFromEnv,
		EnvVar:  "AR_CRED_HELPER_ACCESS_TOKEN_FROM_ENV",
		Example: "AR_CRED_HELPER_ACCESS_TOKEN",
	})

	sec.StringVar(&cli.StringVar{
		Name:   "json-key",
		Usage:  "The path to the JSON key of a service account used for authentication. Setting this flag will result in using JSON key for authentication instead of access token.",
		Target: &f.jsonKeyPath,
		EnvVar: "AR_CRED_HELPER_JSON_KEY",
	})

	sec.DurationVar(&cli.DurationVar{
		Name:    "background-refresh-interval",
		Usage:   "If set, the program will keep running and refresh the credential per given interval, in which case it's best to put this program into background. Recommended value: 5m.",
		Target:  &f.backgroundRefreshInterval,
		EnvVar:  "AR_CRED_HELPER_BACKGROUND_REFRESH_INTERVAL",
		Example: "5m",
	})

	sec.DurationVar(&cli.DurationVar{
		Name:    "background-refresh-duration",
		Usage:   "How long the background refresh will run. The program will exit after this duration or as soon as it fails to refresh the credential.",
		Target:  &f.backgroundRefreshDuration,
		Default: 12 * time.Hour,
		EnvVar:  "AR_CRED_HELPER_BACKGROUND_REFRESH_DURATION",
		Example: "12h",
	})

	return set
}
