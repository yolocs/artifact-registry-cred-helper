package netrc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/abcxyz/pkg/testutil"
)

func TestNetRC_SetToken(t *testing.T) {
	t.Parallel()

	n := &NetRC{content: "existing content\n"}
	hosts := []string{"host1.pkg.dev", "host2.pkg.dev"}
	token := "test-token"

	n.SetToken(hosts, token)

	expected := "existing content\n" +
		"\nmachine host1.pkg.dev\nlogin oauth2accesstoken\npassword test-token\n" +
		"\nmachine host2.pkg.dev\nlogin oauth2accesstoken\npassword test-token\n"

	if n.content != expected {
		t.Errorf("SetToken() content = %q, want %q", n.content, expected)
	}
}

func TestNetRC_SetJSONKey(t *testing.T) {
	t.Parallel()

	n := &NetRC{content: "existing content\n"}
	hosts := []string{"host1.pkg.dev", "host2.pkg.dev"}
	key := "base64-encoded-key"

	n.SetJSONKey(hosts, key)

	expected := "existing content\n" +
		"\nmachine host1.pkg.dev\nlogin _json_key_base64\npassword base64-encoded-key\n" +
		"\nmachine host2.pkg.dev\nlogin _json_key_base64\npassword base64-encoded-key\n"

	if n.content != expected {
		t.Errorf("SetJSONKey() content = %q, want %q", n.content, expected)
	}
}

func TestNetRC_Refresh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  string
		token    string
		expected string
	}{
		{
			name:     "no_existing_token",
			content:  "some content\n",
			token:    "new-token",
			expected: "some content\n",
		},
		{
			name:     "single_token",
			content:  "\nmachine test.pkg.dev\nlogin oauth2accesstoken\npassword old-token\n",
			token:    "new-token",
			expected: "\nmachine test.pkg.dev\nlogin oauth2accesstoken\npassword new-token\n",
		},
		{
			name: "multiple_tokens",
			content: "\nmachine test1.pkg.dev\nlogin oauth2accesstoken\npassword old-token1\n" +
				"\nmachine test2.pkg.dev\nlogin oauth2accesstoken\npassword old-token2\n",
			token: "new-token",
			expected: "\nmachine test1.pkg.dev\nlogin oauth2accesstoken\npassword new-token\n" +
				"\nmachine test2.pkg.dev\nlogin oauth2accesstoken\npassword new-token\n",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			n := &NetRC{content: tc.content}
			n.Refresh(tc.token)
			if n.content != tc.expected {
				t.Errorf("Refresh() content = %q, want %q", n.content, tc.expected)
			}
		})
	}
}

func TestNetRC_Close(t *testing.T) {
	t.Parallel()

	t.Run("successful_write", func(t *testing.T) {
		t.Parallel()

		tempDir, err := os.MkdirTemp("", "netrc-test-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		netrcPath := filepath.Join(tempDir, "test.netrc")
		expectedContent := "test content"

		n := &NetRC{path: netrcPath, content: expectedContent}
		if err := n.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		content, err := os.ReadFile(netrcPath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(content) != expectedContent {
			t.Errorf("Close() wrote %q, want %q", string(content), expectedContent)
		}
	})

	t.Run("directory_creation", func(t *testing.T) {
		t.Parallel()

		tempDir, err := os.MkdirTemp("", "netrc-test-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		netrcPath := filepath.Join(tempDir, "subdir", "test.netrc")
		n := &NetRC{path: netrcPath, content: "test"}

		if err := n.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		if _, err := os.Stat(filepath.Join(tempDir, "subdir")); os.IsNotExist(err) {
			t.Errorf("Close() did not create directory")
		}
	})

	t.Run("directory_creation_error", func(t *testing.T) {
		t.Parallel()

		// Create a file to make the directory creation fail
		tempDir, err := os.MkdirTemp("", "netrc-test-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		// Create a file with the same name as the directory we'll try to create
		dirPath := filepath.Join(tempDir, "file-not-dir")
		if err := os.WriteFile(dirPath, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		netrcPath := filepath.Join(dirPath, "test.netrc")
		n := &NetRC{path: netrcPath, content: "test"}

		err = n.Close()
		if diff := testutil.DiffErrString(err, "failed to create dir"); diff != "" {
			t.Errorf("Close() error diff (-got +want):\n%s", diff)
		}
	})
}

func TestNetRC_update(t *testing.T) {
	t.Parallel()

	t.Run("append_mode", func(t *testing.T) {
		t.Parallel()

		n := &NetRC{content: "existing content\n\nmachine existing.pkg.dev\nlogin oauth2accesstoken\npassword token\n"}
		hosts := []string{"new.pkg.dev"}
		formatter := func(host, pwd string) string {
			return "\nmachine " + host + "\nlogin test\npassword " + pwd + "\n"
		}

		n.update(hosts, formatter, "pwd", true)

		expected := "existing content\n\nmachine existing.pkg.dev\nlogin oauth2accesstoken\npassword token\n" +
			"\nmachine new.pkg.dev\nlogin test\npassword pwd\n"

		if n.content != expected {
			t.Errorf("update() with append content = %q, want %q", n.content, expected)
		}
	})

	t.Run("replace_mode", func(t *testing.T) {
		t.Parallel()

		n := &NetRC{content: "existing content\n\nmachine existing.pkg.dev\nlogin oauth2accesstoken\npassword token\n"}
		hosts := []string{"new.pkg.dev"}
		formatter := func(host, pwd string) string {
			return "\nmachine " + host + "\nlogin test\npassword " + pwd + "\n"
		}

		n.update(hosts, formatter, "pwd", false)

		expected := "existing content\n\nmachine new.pkg.dev\nlogin test\npassword pwd\n"

		if n.content != expected {
			t.Errorf("update() without append content = %q, want %q", n.content, expected)
		}
	})

	t.Run("multiple_hosts", func(t *testing.T) {
		t.Parallel()

		n := &NetRC{content: "existing content\n"}
		hosts := []string{"host1.pkg.dev", "host2.pkg.dev"}
		formatter := func(host, pwd string) string {
			return "\nmachine " + host + "\nlogin test\npassword " + pwd + "\n"
		}

		n.update(hosts, formatter, "pwd", false)

		expected := "existing content\n" +
			"\nmachine host1.pkg.dev\nlogin test\npassword pwd\n" +
			"\nmachine host2.pkg.dev\nlogin test\npassword pwd\n"

		if n.content != expected {
			t.Errorf("update() with multiple hosts content = %q, want %q", n.content, expected)
		}
	})
}
