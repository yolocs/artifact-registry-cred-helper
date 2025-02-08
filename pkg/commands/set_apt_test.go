package commands

import (
	"context"
	"testing"
)

func TestSetAptCommand_runOnce(t *testing.T) {
	tests := []struct {
		name        string
		command     *SetAptCommand
		mockAuth    *mockAuthConfig
		wantToken   string
		wantJSONKey string
		wantHosts   []string
		wantErr     bool
		setEnv      map[string]string
	}{
		{
			name: "get auth token success",
			command: &SetAptCommand{
				baseCommand: baseCommand{
					getAuthToken: func(context.Context) (string, error) {
						return "test-token", nil
					},
				},
				commonFlags: &CommonFlags{
					repoURLs: []string{"us-apt.pkg.dev/proj/repo"},
				},
			},
			mockAuth:  &mockAuthConfig{},
			wantToken: "test-token",
			wantHosts: []string{"us-apt.pkg.dev"},
		},
		{
			name: "get json key success",
			command: &SetAptCommand{
				baseCommand: baseCommand{
					getEncodedJSONKey: func(string) (string, error) {
						return "encoded-key", nil
					},
				},
				commonFlags: &CommonFlags{
					repoURLs:    []string{"us-apt.pkg.dev/proj/repo"},
					jsonKeyPath: "/path/to/key.json",
				},
			},
			mockAuth:    &mockAuthConfig{},
			wantJSONKey: "encoded-key",
			wantHosts:   []string{"us-apt.pkg.dev"},
		},
		{
			name: "get token from env success",
			command: &SetAptCommand{
				commonFlags: &CommonFlags{
					repoURLs:           []string{"us-apt.pkg.dev/proj/repo"},
					accessTokenFromEnv: "TEST_TOKEN",
				},
			},
			mockAuth:  &mockAuthConfig{},
			setEnv:    map[string]string{"TEST_TOKEN": "env-token"},
			wantToken: "env-token",
			wantHosts: []string{"us-apt.pkg.dev"},
		},
		{
			name: "get token from env failure - env not set",
			command: &SetAptCommand{
				commonFlags: &CommonFlags{
					repoURLs:           []string{"us-apt.pkg.dev/proj/repo"},
					accessTokenFromEnv: "TEST_TOKEN",
				},
			},
			mockAuth: &mockAuthConfig{},
			wantErr:  true,
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
				if len(tc.wantHosts) != len(tc.mockAuth.hosts) {
					t.Errorf("hosts = %v, want %v", tc.mockAuth.hosts, tc.wantHosts)
				}
				if !tc.mockAuth.closed {
					t.Error("config was not closed")
				}
			}
		})
	}
}
