package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetCommand_Run(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		stdin         string
		getAuthToken  authTokenGetter
		expectedHosts []string
		expectedErr   error
		expectedOut   string
	}{
		{
			name: "valid hosts flag",
			args: []string{"--hosts", "us-go.pkg.dev"},
			getAuthToken: func(ctx context.Context) (string, error) {
				return "test-token", nil
			},
			expectedHosts: []string{"us-go.pkg.dev"},
			expectedOut:   `{"headers":{"Authorization":["Bearer test-token"]}}`,
		},
		{
			name:  "valid stdin",
			stdin: `{"uri": "us-go.pkg.dev"}`,
			getAuthToken: func(ctx context.Context) (string, error) {
				return "test-token", nil
			},
			expectedHosts: []string{"us-go.pkg.dev"},
			expectedOut:   `{"headers":{"Authorization":["Bearer test-token"]}}`,
		},
		{
			name: "invalid host",
			args: []string{"--hosts", "invalid-host"},
			getAuthToken: func(ctx context.Context) (string, error) {
				return "test-token", nil
			},
			expectedErr: errors.New("host \"invalid-host\" doesn't have domain '.pkg.dev'"),
		},
		{
			name: "getAuthToken error",
			args: []string{"--hosts", "us-go.pkg.dev"},
			getAuthToken: func(ctx context.Context) (string, error) {
				return "", errors.New("token error")
			},
			expectedErr: errors.New("failed to get access token: token error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &GetCommand{
				baseCommand: baseCommand{
					getAuthToken: tt.getAuthToken,
				},
			}

			var stdin bytes.Buffer
			stdin.WriteString(tt.stdin)
			cmd.SetStdin(&stdin)

			var stdout bytes.Buffer
			cmd.SetStdout(&stdout)

			err := cmd.Run(context.Background(), tt.args)
			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectedOut != "" {
				var got map[string]interface{}
				if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
					t.Fatalf("failed to unmarshal output: %v", err)
				}

				var expected map[string]interface{}
				if err := json.Unmarshal([]byte(tt.expectedOut), &expected); err != nil {
					t.Fatalf("failed to unmarshal expected output: %v", err)
				}

				if !cmp.Equal(got, expected) {
					t.Fatalf("expected output %v, got %v", expected, got)
				}
			}
		})
	}
}
