package commands

import (
	"testing"
	"time"

	"github.com/abcxyz/pkg/testutil"
)

func TestCommonFlags_validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		flags   *CommonFlags
		wantErr string
	}{
		{
			name: "valid config",
			flags: &CommonFlags{
				repoURLs:                  []string{"us-go.pkg.dev/my-project/repo1"},
				backgroundRefreshInterval: 5 * time.Minute,
			},
			wantErr: "",
		},
		{
			name: "no hosts specified",
			flags: &CommonFlags{
				backgroundRefreshInterval: 5 * time.Minute,
			},
			wantErr: "no host specified",
		},
		{
			name: "invalid url format",
			flags: &CommonFlags{
				repoURLs:                  []string{"invalid-url"},
				backgroundRefreshInterval: 5 * time.Minute,
			},
			wantErr: `repo URL "https://invalid-url" not in format '*.pkg.dev/[project]/[repo]'`,
		},
		{
			name: "both json key and access token set",
			flags: &CommonFlags{
				repoURLs:           []string{"us-go.pkg.dev/my-project/repo1"},
				jsonKeyPath:        "/path/to/key.json",
				accessTokenFromEnv: "TOKEN",
			},
			wantErr: "only one of --json-key or --access-token-from-env can be set",
		},
		{
			name: "refresh interval too short",
			flags: &CommonFlags{
				repoURLs:                  []string{"us-go.pkg.dev/my-project/repo1"},
				backgroundRefreshInterval: 1 * time.Minute,
			},
			wantErr: "background refresh interval must be at least 2 minutes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.flags.validate()
			if diff := testutil.DiffErrString(err, tt.wantErr); diff != "" {
				t.Errorf("CommonFlags.validate() %s", diff)
			}
		})
	}
}

func TestCommonFlags_validateWithoutURLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		flags   *CommonFlags
		wantErr string
	}{
		{
			name: "valid config",
			flags: &CommonFlags{
				backgroundRefreshInterval: 5 * time.Minute,
			},
			wantErr: "",
		},
		{
			name: "refresh interval too short",
			flags: &CommonFlags{
				backgroundRefreshInterval: 1 * time.Minute,
			},
			wantErr: "background refresh interval must be at least 2 minutes",
		},
		{
			name: "both auth methods set",
			flags: &CommonFlags{
				jsonKeyPath:        "/path/to/key.json",
				accessTokenFromEnv: "TOKEN",
			},
			wantErr: "only one of --json-key or --access-token-from-env can be set",
		},
		{
			name: "zero refresh interval",
			flags: &CommonFlags{
				backgroundRefreshInterval: 0,
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.flags.validateWithoutURLs()
			if diff := testutil.DiffErrString(err, tt.wantErr); diff != "" {
				t.Errorf("CommonFlags.validateWithoutURLs() %s", diff)
			}
		})
	}
}
