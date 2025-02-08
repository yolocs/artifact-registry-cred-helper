package maven

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDefaultRepoID(t *testing.T) {
	testCases := []struct {
		repoURL  string
		expected string
	}{
		{
			repoURL:  "https://us-maven.pkg.dev/my-project/my-repo",
			expected: "artifactregistry-my-project-my-repo",
		},
		{
			repoURL:  "https://europe-maven.pkg.dev/another-project/another-repo",
			expected: "artifactregistry-another-project-another-repo",
		},
		{
			repoURL:  "https://asia-maven.pkg.dev/yet-another-project/yet-another-repo",
			expected: "artifactregistry-yet-another-project-yet-another-repo",
		},
	}

	for _, tc := range testCases {
		u, err := url.Parse(tc.repoURL)
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}
		actual := DefaultRepoID(u)
		if actual != tc.expected {
			t.Errorf("DefaultRepoID(%q) = %q, expected %q", tc.repoURL, actual, tc.expected)
		}
	}
}

func TestSettings_SetToken(t *testing.T) {
	settingsPath := filepath.Join(t.TempDir(), "settings.xml")
	settings, err := Open(settingsPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() {
		if err := settings.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	repoIDs := []string{"repo1", "repo2"}
	token := "test-token"
	settings.SetToken(repoIDs, token)

	doc := settings.doc
	for _, repoID := range repoIDs {
		server := doc.FindElement(strings.ReplaceAll("//settings/servers/server[id='"+repoID+"']", "'", "\""))
		if server == nil {
			t.Fatalf("server with id %q not found", repoID)
		}
		username := server.FindElement("username")
		if username == nil {
			t.Fatalf("username for server %q not found", repoID)
		}
		if username.Text() != "oauth2accesstoken" {
			t.Errorf("username = %q, expected %q", username.Text(), "oauth2accesstoken")
		}
		password := server.FindElement("password")
		if password == nil {
			t.Fatalf("password for server %q not found", repoID)
		}
		if password.Text() != token {
			t.Errorf("password = %q, expected %q", password.Text(), token)
		}
	}
}

func TestSettings_SetJSONKey(t *testing.T) {
	settingsPath := filepath.Join(t.TempDir(), "settings.xml")
	settings, err := Open(settingsPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() {
		if err := settings.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	repoIDs := []string{"repo1", "repo2"}
	jsonKey := "test-json-key"
	settings.SetJSONKey(repoIDs, jsonKey)

	doc := settings.doc
	for _, repoID := range repoIDs {
		server := doc.FindElement(strings.ReplaceAll("//settings/servers/server[id='"+repoID+"']", "'", "\""))
		if server == nil {
			t.Fatalf("server with id %q not found", repoID)
		}
		username := server.FindElement("username")
		if username == nil {
			t.Fatalf("username for server %q not found", repoID)
		}
		if username.Text() != "_json_key_base64" {
			t.Errorf("username = %q, expected %q", username.Text(), "_json_key_base64")
		}
		password := server.FindElement("password")
		if password == nil {
			t.Fatalf("password for server %q not found", repoID)
		}
		if password.Text() != jsonKey {
			t.Errorf("password = %q, expected %q", password.Text(), jsonKey)
		}
	}
}

func TestSettings_update(t *testing.T) {
	settingsPath := filepath.Join(t.TempDir(), "settings.xml")
	settings, err := Open(settingsPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() {
		if err := settings.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	repoIDs := []string{"repo1", "repo2"}
	user := "test-user"
	pwd := "test-password"
	settings.update(repoIDs, user, pwd)

	doc := settings.doc
	for _, repoID := range repoIDs {
		server := doc.FindElement(strings.ReplaceAll("//settings/servers/server[id='"+repoID+"']", "'", "\""))
		if server == nil {
			t.Fatalf("server with id %q not found", repoID)
		}
		id := server.FindElement("id")
		if id == nil {
			t.Fatalf("id for server %q not found", repoID)
		}
		if id.Text() != repoID {
			t.Errorf("id = %q, expected %q", id.Text(), repoID)
		}
		username := server.FindElement("username")
		if username == nil {
			t.Fatalf("username for server %q not found", repoID)
		}
		if username.Text() != user {
			t.Errorf("username = %q, expected %q", username.Text(), user)
		}
		password := server.FindElement("password")
		if password == nil {
			t.Fatalf("password for server %q not found", repoID)
		}
		if password.Text() != pwd {
			t.Errorf("password = %q, expected %q", password.Text(), pwd)
		}
	}

	// Update existing server
	newUser := "new-test-user"
	newPwd := "new-test-password"
	settings.update(repoIDs, newUser, newPwd)

	for _, repoID := range repoIDs {
		server := doc.FindElement(strings.ReplaceAll("//settings/servers/server[id='"+repoID+"']", "'", "\""))
		if server == nil {
			t.Fatalf("server with id %q not found", repoID)
		}
		username := server.FindElement("username")
		if username == nil {
			t.Fatalf("username for server %q not found", repoID)
		}
		if username.Text() != newUser {
			t.Errorf("username = %q, expected %q", username.Text(), newUser)
		}
		password := server.FindElement("password")
		if password == nil {
			t.Fatalf("password for server %q not found", repoID)
		}
		if password.Text() != newPwd {
			t.Errorf("password = %q, expected %q", password.Text(), newPwd)
		}
	}
}

func TestSettings_Close(t *testing.T) {
	t.Run("create directory", func(t *testing.T) {
		settingsPath := filepath.Join(t.TempDir(), "nested", "settings.xml")
		settings, err := Open(settingsPath)
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}

		err = settings.Close()
		if err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		if _, err := os.Stat(settingsPath); err != nil {
			t.Errorf("settings.xml should exist, but got error: %v", err)
		}
	})

	t.Run("write to file", func(t *testing.T) {
		settingsPath := filepath.Join(t.TempDir(), "settings.xml")
		settings, err := Open(settingsPath)
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}

		repoIDs := []string{"repo1"}
		user := "test-user"
		pwd := "test-password"
		settings.update(repoIDs, user, pwd)

		err = settings.Close()
		if err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		content, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("failed to read settings.xml: %v", err)
		}

		expectedContent := `<settings xmlns="http://maven.apache.org/SETTINGS/1.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.0.0 http://maven.apache.org/xsd/settings-1.0.0.xsd">
  <servers>
    <server>
      <id>repo1</id>
      <username>test-user</username>
      <password>test-password</password>
    </server>
  </servers>
</settings>
`
		if diff := cmp.Diff(string(content), expectedContent); diff != "" {
			t.Errorf("settings.xml content mismatch (-got +want):\n%s", diff)
		}
	})
}
