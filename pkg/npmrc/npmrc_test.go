package npmrc

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abcxyz/pkg/testutil"
)

func TestOpenNonExistingFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	npmrcPath := filepath.Join(tmpDir, "nonexistent.npmrc")

	// Should succeed even if file doesn't exist.
	cfg, err := Open(npmrcPath, "")
	if diff := testutil.DiffErrString(err, ""); diff != "" {
		t.Errorf("Open unexpected error: %s", diff)
	}

	// Since file did not exist, content should be empty.
	if cfg.content.Len() != 0 {
		t.Errorf("expected empty content for new npmrc, got: %q", cfg.content.String())
	}
}

func TestSetToken(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	npmrcPath := filepath.Join(tmpDir, "test.npmrc")

	// Create an empty file.
	if err := os.WriteFile(npmrcPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Open the file.
	cfg, err := Open(npmrcPath, "")
	if diff := testutil.DiffErrString(err, ""); diff != "" {
		t.Errorf("Open unexpected error: %s", diff)
	}

	repo := "us-npm.pkg.dev/my-project/repo1"
	token := "testtoken"

	cfg.SetToken([]string{repo}, token)

	// Expected values.
	url := normalizeRepoURL(repo)
	// Since scope is empty, registry key is "registry"
	registryLine := "registry=" + url
	// Registry derivation: remove "https:" prefix.
	registry := strings.TrimPrefix(url, "https:")
	authLine := registry + ":always-auth=true"
	emailLine := registry + ":email=not.valid@email.com"
	encoded := base64.StdEncoding.EncodeToString([]byte("oauth2accesstoken:" + token))
	tokenLine := registry + ":_authToken=" + encoded

	expected := []string{
		registryLine,
		authLine,
		emailLine,
		tokenLine,
	}

	content := cfg.content.String()
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) != len(expected) {
		t.Fatalf("expected %d lines got %d: %q", len(expected), len(lines), content)
	}

	// Check each expected line.
	for i, exp := range expected {
		if lines[i] != exp {
			t.Errorf("line %d mismatch, expected %q got %q", i+1, exp, lines[i])
		}
	}
}

func TestSetJSONKey(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	npmrcPath := filepath.Join(tmpDir, "test_json.npmrc")

	// Start with some pre-existing content.
	initial := "registry=https://us-npm.pkg.dev/my-project/repo0/\n"
	if err := os.WriteFile(npmrcPath, []byte(initial), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg, err := Open(npmrcPath, "myscope")
	if diff := testutil.DiffErrString(err, ""); diff != "" {
		t.Errorf("Open unexpected error: %s", diff)
	}

	repo := "us-npm.pkg.dev/my-project/repo1"
	base64Key := "base64jsonkey"
	cfg.SetJSONKey([]string{repo}, base64Key)

	url := normalizeRepoURL(repo)
	// With scope set, registry key becomes "@myscope:registry"
	registryLine := "@myscope:registry=" + url
	registry := strings.TrimPrefix(url, "https:")
	authLine := registry + ":always-auth=true"
	emailLine := registry + ":email=not.valid@email.com"
	encoded := base64.StdEncoding.EncodeToString([]byte("_json_key_base64:" + base64Key))
	tokenLine := registry + ":_authToken=" + encoded

	// The file already had one registry line but update will replace it accordingly.
	expectedLines := []string{
		"registry=https://us-npm.pkg.dev/my-project/repo0/",
		registryLine,
		authLine,
		emailLine,
		tokenLine,
	}

	content := cfg.content.String()
	lines := strings.Split(strings.TrimSpace(content), "\n")

	if len(lines) != len(expectedLines) {
		t.Fatalf("expected %d lines but got %d: %q", len(expectedLines), len(lines), content)
	}

	for i, exp := range expectedLines {
		if lines[i] != exp {
			t.Errorf("line %d mismatch, expected %q got %q", i+1, exp, lines[i])
		}
	}
}

func TestCloseWritesFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	npmrcPath := filepath.Join(tmpDir, "close_test.npmrc")

	cfg, err := Open(npmrcPath, "")
	if diff := testutil.DiffErrString(err, ""); diff != "" {
		t.Errorf("Open unexpected error: %s", diff)
	}

	repo := "us-npm.pkg.dev/my-project/repo2"
	token := "closetesttoken"
	cfg.SetToken([]string{repo}, token)

	if err := cfg.Close(); err != nil {
		if diff := testutil.DiffErrString(err, ""); diff != "" {
			t.Errorf("Close unexpected error: %s", diff)
		}
	}

	// Re-read file to verify content.
	data, err := os.ReadFile(npmrcPath)
	if diff := testutil.DiffErrString(err, ""); diff != "" {
		t.Errorf("ReadFile unexpected error: %s", diff)
	}

	url := normalizeRepoURL(repo)
	registryLine := "registry=" + url
	registry := strings.TrimPrefix(url, "https:")
	authLine := registry + ":always-auth=true"
	emailLine := registry + ":email=not.valid@email.com"
	encoded := base64.StdEncoding.EncodeToString([]byte("oauth2accesstoken:" + token))
	tokenLine := registry + ":_authToken=" + encoded

	expected := []string{
		registryLine,
		authLine,
		emailLine,
		tokenLine,
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != len(expected) {
		t.Fatalf("expected %d lines from file but got %d: %q", len(expected), len(lines), data)
	}
	for i, exp := range expected {
		if lines[i] != exp {
			t.Errorf("line %d mismatch, expected %q got %q", i+1, exp, lines[i])
		}
	}
}
