package commands

import (
	"context"
	"net/url"
	"testing"

	"github.com/abcxyz/pkg/testutil"
)

func TestSetMavenSettings_runOnce(t *testing.T) {
	tests := []struct {
		name        string
		command     *SetMavenCommand
		mockAuth    *mockAuthConfig
		wantToken   string
		wantJSONKey string
		wantRepoIDs []string
		wantErr     string
		setEnv      map[string]string
	}{
		{
			name: "get auth token success",
			command: &SetMavenCommand{
				baseCommand: baseCommand{
					getAuthToken: func(context.Context) (string, error) {
						return "test-token", nil
					},
				},
				commonFlags: &CommonFlags{
					parsedURLs: []*url.URL{{Host: "us-maven.pkg.dev", Path: "/proj/repo"}},
				},
			},
			mockAuth:    &mockAuthConfig{},
			wantToken:   "test-token",
			wantRepoIDs: []string{"artifactregistry-proj-repo"},
		},
		{
			name: "get json key success",
			command: &SetMavenCommand{
				baseCommand: baseCommand{
					getEncodedJSONKey: func(string) (string, error) {
						return "encoded-key", nil
					},
				},
				commonFlags: &CommonFlags{
					parsedURLs:  []*url.URL{{Host: "us-maven.pkg.dev", Path: "/proj/repo"}},
					jsonKeyPath: "/path/to/key.json",
				},
			},
			mockAuth:    &mockAuthConfig{},
			wantJSONKey: "encoded-key",
			wantRepoIDs: []string{"artifactregistry-proj-repo"},
		},
		{
			name: "get token from env success",
			command: &SetMavenCommand{
				commonFlags: &CommonFlags{
					parsedURLs:         []*url.URL{{Host: "us-maven.pkg.dev", Path: "/proj/repo"}},
					accessTokenFromEnv: "TEST_TOKEN",
				},
			},
			mockAuth:    &mockAuthConfig{},
			setEnv:      map[string]string{"TEST_TOKEN": "env-token"},
			wantToken:   "env-token",
			wantRepoIDs: []string{"artifactregistry-proj-repo"},
		},
		{
			name: "get token from env failure - env not set",
			command: &SetMavenCommand{
				commonFlags: &CommonFlags{
					repoURLs:           []string{"us-maven.pkg.dev/proj/repo"},
					accessTokenFromEnv: "TEST_TOKEN",
				},
			},
			mockAuth: &mockAuthConfig{},
			wantErr:  "failed to get access token from env var",
		},
		{
			name: "override repo IDs success",
			command: &SetMavenCommand{
				baseCommand: baseCommand{
					getAuthToken: func(context.Context) (string, error) {
						return "test-token", nil
					},
				},
				commonFlags: &CommonFlags{
					repoURLs: []string{"us-maven.pkg.dev/proj/repo"},
				},
				repoIDsOverride: []string{"custom-repo-id"},
			},
			mockAuth:    &mockAuthConfig{},
			wantToken:   "test-token",
			wantRepoIDs: []string{"custom-repo-id"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.setEnv {
				t.Setenv(k, v)
			}
			err := tc.command.runOnce(context.Background(), tc.mockAuth)
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf("runOnce() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if tc.wantErr == "" {
				if tc.wantToken != tc.mockAuth.token {
					t.Errorf("token = %v, want %v", tc.mockAuth.token, tc.wantToken)
				}
				if tc.wantJSONKey != tc.mockAuth.jsonKey {
					t.Errorf("jsonKey = %v, want %v", tc.mockAuth.jsonKey, tc.wantJSONKey)
				}
				if len(tc.wantRepoIDs) != len(tc.mockAuth.hosts) {
					t.Errorf("repoIDs = %v, want %v", tc.mockAuth.hosts, tc.wantRepoIDs)
				}
				if !tc.mockAuth.closed {
					t.Error("config was not closed")
				}
			}
		})
	}
}
