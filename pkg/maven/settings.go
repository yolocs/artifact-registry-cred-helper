package maven

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/beevik/etree"
)

// DefaultRepoID is our default way to construct repo ID used in pom.xml and
// settings.xml.
//
// Given repo URL: us-maven.pkg.dev/my-project/my-repo
// Default repo ID: artifactregistry-my-project-my-repo
func DefaultRepoID(repoURL *url.URL) string {
	return "artifactregistry" + strings.ReplaceAll(repoURL.Path, "/", "-")
}

type Settings struct {
	path string
	doc  *etree.Document
}

func Open(settingsPath string) (*Settings, error) {
	if settingsPath == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot open .netrc file: %w", err)
		}
		settingsPath = path.Join(h, ".m2")
	}

	if !strings.HasSuffix(settingsPath, "settings.xml") {
		settingsPath = path.Join(settingsPath, "settings.xml")
	}

	doc := etree.NewDocument()
	_, err := os.Stat(settingsPath)
	if err != nil && !os.IsNotExist(err) { // Handle other errors.
		return nil, fmt.Errorf("failed to stat Maven settings.xml file: %w", err)
	}

	if err == nil { // File exists.
		if err := doc.ReadFromFile(settingsPath); err != nil {
			return nil, fmt.Errorf("cannot load Maven settings.xml file at %q: %w", settingsPath, err)
		}
	} else {
		// Start a new settings file.
		settings := doc.CreateElement("settings")
		settings.CreateAttr("xmlns", "http://maven.apache.org/SETTINGS/1.0.0")
		settings.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
		settings.CreateAttr("xsi:schemaLocation", "http://maven.apache.org/SETTINGS/1.0.0 http://maven.apache.org/xsd/settings-1.0.0.xsd")
	}

	return &Settings{path: settingsPath, doc: doc}, nil
}

func (s *Settings) Close() error {
	// mkdir for the settings file since etree doesn't handle that.
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return fmt.Errorf("failed to create Maven settings.xml directory: %w", err)
	}

	s.doc.Indent(2) // Make it pretty.
	if err := s.doc.WriteToFile(s.path); err != nil {
		return fmt.Errorf("failed to save Maven settings.xml at %q: %w", s.path, err)
	}

	return nil
}

func (s *Settings) SetToken(repoIDs []string, token string) {
	s.update(repoIDs, "oauth2accesstoken", token)
}

func (s *Settings) SetJSONKey(repoIDs []string, base64Key string) {
	s.update(repoIDs, "_json_key_base64", base64Key)
}

func (s *Settings) update(repoIDs []string, user, pwd string) {
	servers := s.doc.FindElement("//settings/servers")
	if servers == nil { // Create servers if it doesn't exist
		servers = s.doc.CreateElement("servers")
		s.doc.FindElement("//settings").AddChild(servers)
	}

	for _, repoID := range repoIDs {
		existingServer := false
		for _, server := range servers.ChildElements() {
			idElem := server.FindElement("id")
			if idElem != nil && idElem.Text() == repoID {
				username := server.FindElement("username")
				if username == nil {
					username = server.CreateElement("username")
				}
				username.SetText(user)
				password := server.FindElement("password")
				if password == nil {
					password = server.CreateElement("password")
				}
				password.SetText(pwd)

				existingServer = true
				break
			}
		}

		if !existingServer {
			newServer := servers.CreateElement("server")
			newServer.CreateElement("id").SetText(repoID)
			newServer.CreateElement("username").SetText(user)
			newServer.CreateElement("password").SetText(pwd)
		}
	}
}
