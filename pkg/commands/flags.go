package commands

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/abcxyz/pkg/cli"
)

type CommonFlags struct {
	repoURLs                  []string
	accessTokenFromEnv        string
	jsonKeyPath               string
	backgroundRefreshInterval time.Duration
	backgroundRefreshDuration time.Duration

	parsedURLs []*url.URL
	once       sync.Once
}

func (f *CommonFlags) validate() error {
	var merr error

	if err := f.validateWithoutURLs(); err != nil {
		merr = errors.Join(merr, err)
	}

	if len(f.repoURLs) <= 0 {
		merr = errors.Join(merr, fmt.Errorf("no host specified"))
	}

	if err := f.parseURLs(); err != nil {
		merr = errors.Join(merr, err)
	}

	return merr
}

func (f *CommonFlags) validateWithoutURLs() error {
	var merr error

	if f.backgroundRefreshInterval < 2*time.Minute && f.backgroundRefreshInterval > 0 {
		merr = errors.Join(merr, fmt.Errorf("background refresh interval must be at least 2 minutes"))
	}

	if f.jsonKeyPath != "" && f.accessTokenFromEnv != "" {
		merr = errors.Join(merr, fmt.Errorf("only one of --json-key or --access-token-from-env can be set"))
	}

	return merr
}

// If validate was called, then should return no error.
func (f *CommonFlags) repoHosts() ([]string, error) {
	if err := f.parseURLs(); err != nil {
		return nil, err
	}

	set := map[string]struct{}{}
	hosts := make([]string, 0, len(f.parsedURLs))
	for _, u := range f.parsedURLs {
		if _, ok := set[u.Host]; ok {
			continue
		}
		set[u.Host] = struct{}{}
		hosts = append(hosts, u.Host)
	}

	return hosts, nil
}

func (f *CommonFlags) parseURLs() (merr error) {
	f.once.Do(func() {
		for _, h := range f.repoURLs {
			if !strings.HasPrefix(h, "https://") {
				h = "https://" + h
			}
			u, err := url.Parse(h)
			if err != nil {
				merr = errors.Join(merr, fmt.Errorf("failed to parse host %q: %w", h, err))
				continue
			}
			if !strings.HasSuffix(u.Host, ".pkg.dev") || len(strings.Split(strings.Trim(u.Path, "/"), "/")) != 2 {
				merr = errors.Join(merr, fmt.Errorf("repo URL %q not in format '*.pkg.dev/[project]/[repo]'", u.String()))
				continue
			}
			f.parsedURLs = append(f.parsedURLs, u)
		}
	})

	return
}

func (f *CommonFlags) setSection(set *cli.FlagSet) *cli.FlagSet {
	sec := set.NewSection("COMMON OPTIONS")

	sec.StringSliceVar(&cli.StringSliceVar{
		Name:    "repo-urls",
		Usage:   "REQUIRED. The hosts to set the credential for. Must in format '*.pkg.dev/[project]/[repo]'.",
		Target:  &f.repoURLs,
		EnvVar:  "AR_CRED_HELPER_HOSTS",
		Example: "us-go.pkg.dev/my-project/repo1,asia-maven.pkg.dev/my-project/repo2",
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
