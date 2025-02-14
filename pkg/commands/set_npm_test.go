package commands

import (
	"context"
	"testing"
)

func TestSetNPMCommand_runOnce(t *testing.T) {
	tests := []struct {
		name         string
		command      *SetNPMCommand
		mockAuth     *mockAuthConfig
		wantToken    string
		wantJSONKey  string
		wantRepoURLs []string
		wantErr      bool
		setEnv       map[string]string
	}{
		{
			name: "get JSON key success",
			command: &SetNPMCommand{
				baseCommand: baseCommand{
					getEncodedJSONKey: func(path string) (string, error) {
						return "encoded-json-key", nil
					},
				},
				commonFlags: &CommonFlags{
					repoURLs:    []string{"eu.pkg.dev/proj/repo"},
					jsonKeyPath: "/path/to/key.json",
				},
			},
			mockAuth:     &mockAuthConfig{},
			wantJSONKey:  "encoded-json-key",
			wantRepoURLs: []string{"eu.pkg.dev/proj/repo"},
		},
		{
			name: "get token from env success",
			command: &SetNPMCommand{
				commonFlags: &CommonFlags{
					repoURLs:           []string{"eu.pkg.dev/proj/repo"},
					accessTokenFromEnv: "TEST_NPM_TOKEN",
				},
			},
			mockAuth:     &mockAuthConfig{},
			setEnv:       map[string]string{"TEST_NPM_TOKEN": "env-token"},
			wantToken:    "env-token",
			wantRepoURLs: []string{"eu.pkg.dev/proj/repo"},
		},
		{
			name: "get token from env failure - env not set",
			command: &SetNPMCommand{
				commonFlags: &CommonFlags{
					repoURLs:           []string{"eu.pkg.dev/proj/repo"},
					accessTokenFromEnv: "TEST_NPM_TOKEN",
				},
			},
			mockAuth: &mockAuthConfig{},
			wantErr:  true,
		},
		{
			name: "get auth token success",
			command: &SetNPMCommand{
				baseCommand: baseCommand{
					getAuthToken: func(ctx context.Context) (string, error) {
						return "auth-token", nil
					},
				},
				commonFlags: &CommonFlags{
					repoURLs: []string{"eu.pkg.dev/proj/repo"},
				},
			},
			mockAuth:     &mockAuthConfig{},
			wantToken:    "auth-token",
			wantRepoURLs: []string{"eu.pkg.dev/proj/repo"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.setEnv {
				t.Setenv(k, v)
			}
			err := tc.command.runOnce(context.Background(), tc.mockAuth)
			if (err != nil) != tc.wantErr {
				t.Errorf("runOnce() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				if tc.wantToken != tc.mockAuth.token {
					t.Errorf("token = %v, want %v", tc.mockAuth.token, tc.wantToken)
				}
				if tc.wantJSONKey != tc.mockAuth.jsonKey {
					t.Errorf("jsonKey = %v, want %v", tc.mockAuth.jsonKey, tc.wantJSONKey)
				}
				if len(tc.wantRepoURLs) != len(tc.mockAuth.hosts) {
					t.Errorf("hosts = %v, want %v", tc.mockAuth.hosts, tc.wantRepoURLs)
				} else {
					for i, h := range tc.wantRepoURLs {
						if h != tc.mockAuth.hosts[i] {
							t.Errorf("host[%d] = %v, want %v", i, tc.mockAuth.hosts[i], h)
						}
					}
				}
				if !tc.mockAuth.closed {
					t.Error("config was not closed")
				}
			}
		})
	}
}
