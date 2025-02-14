package commands

import (
	"testing"
	"time"
)

func TestCommonFlags_validate(t *testing.T) {
	tests := []struct {
		name    string
		flags   *CommonFlags
		wantErr bool
	}{
		{
			name: "valid config",
			flags: &CommonFlags{
				repoURLs:                  []string{"us-go.pkg.dev/my-project/repo1"},
				backgroundRefreshInterval: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "no hosts specified",
			flags: &CommonFlags{
				backgroundRefreshInterval: 5 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid url format",
			flags: &CommonFlags{
				repoURLs:                  []string{"invalid-url"},
				backgroundRefreshInterval: 5 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "both json key and access token set",
			flags: &CommonFlags{
				repoURLs:           []string{"us-go.pkg.dev/my-project/repo1"},
				jsonKeyPath:        "/path/to/key.json",
				accessTokenFromEnv: "TOKEN",
			},
			wantErr: true,
		},
		{
			name: "refresh interval too short",
			flags: &CommonFlags{
				repoURLs:                  []string{"us-go.pkg.dev/my-project/repo1"},
				backgroundRefreshInterval: 1 * time.Minute,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.flags.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CommonFlags.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommonFlags_validateWithoutURLs(t *testing.T) {
	tests := []struct {
		name    string
		flags   *CommonFlags
		wantErr bool
	}{
		{
			name: "valid config",
			flags: &CommonFlags{
				backgroundRefreshInterval: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "refresh interval too short",
			flags: &CommonFlags{
				backgroundRefreshInterval: 1 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "both auth methods set",
			flags: &CommonFlags{
				jsonKeyPath:        "/path/to/key.json",
				accessTokenFromEnv: "TOKEN",
			},
			wantErr: true,
		},
		{
			name: "zero refresh interval",
			flags: &CommonFlags{
				backgroundRefreshInterval: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.flags.validateWithoutURLs()
			if (err != nil) != tt.wantErr {
				t.Errorf("CommonFlags.validateWithoutURLs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
