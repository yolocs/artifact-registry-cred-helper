package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/abcxyz/pkg/testutil"
	"github.com/google/go-cmp/cmp"
)

func TestGetCommand_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          []string
		stdin         string
		getAuthToken  authTokenGetter
		expectedHosts []string
		expectedErr   string
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
			expectedErr: "host \"invalid-host\" doesn't have domain '.pkg.dev'",
		},
		{
			name: "getAuthToken error",
			args: []string{"--hosts", "us-go.pkg.dev"},
			getAuthToken: func(ctx context.Context) (string, error) {
				return "", fmt.Errorf("token error")
			},
			expectedErr: "failed to get access token: token error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
			if diff := testutil.DiffErrString(err, tt.expectedErr); diff != "" {
				t.Errorf("unexpected error: %s", diff)
				return
			}

			if tt.expectedErr != "" {
				return
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
