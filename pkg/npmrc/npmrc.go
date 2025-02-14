// Package npmrc provides functions to modify an npmrc file.
package npmrc

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	npmrcPath string
	scope     string
	content   *bytes.Buffer
}

func Open(npmrcPath, scope string) (*Config, error) {
	if npmrcPath == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot find HOME dir: %w", err)
		}
		npmrcPath = filepath.Join(h, ".npmrc")
	}

	b, err := os.ReadFile(npmrcPath)
	if os.IsNotExist(err) {
		return &Config{npmrcPath: npmrcPath, scope: scope, content: &bytes.Buffer{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot load file %q: %v", npmrcPath, err)
	}

	return &Config{npmrcPath: npmrcPath, scope: scope, content: bytes.NewBuffer(b)}, nil
}

func (c *Config) SetToken(repos []string, token string) {
	c.update(repos, "oauth2accesstoken", token)
}

func (c *Config) SetJSONKey(repos []string, base64Key string) {
	c.update(repos, "_json_key_base64", base64Key)
}

func (c *Config) Close() error {
	// Make sure dir exists.
	if err := os.MkdirAll(filepath.Dir(c.npmrcPath), 0755); err != nil {
		return fmt.Errorf("failed to create dir for %q: %w", c.npmrcPath, err)
	}

	f, err := os.OpenFile(c.npmrcPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		return fmt.Errorf("failed to open %q: %w", c.npmrcPath, err)
	}
	defer f.Close()

	if _, err := f.Write(c.content.Bytes()); err != nil {
		return fmt.Errorf("failed to save %q: %w", c.npmrcPath, err)
	}
	return nil
}

func (c *Config) update(repos []string, user, pwd string) {
	scanner := bufio.NewScanner(c.content)
	var lines []string

	// For each registry, we need 4 lines config:
	// @scope:registry=https://us-go.pkg.dev/my-project/repo1/
	// //us-go.pkg.dev/my-project/repo1/:always-auth=true
	// //us-go.pkg.dev/my-project/repo1/:_authToken=base64-encoded-token
	// //us-go.pkg.dev/my-project/repo1/:email=not.valid@email.com
	existingRegistries := map[string]struct{}{}
	existingCreds := map[string]struct{}{}
	existingAuth := map[string]struct{}{}
	existingEmails := map[string]struct{}{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			lines = append(lines, line)
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			lines = append(lines, line)
			continue // Ignore invalid line.
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		newLine := line

		for _, repo := range repos {
			url := normalizeRepoURL(repo)
			registry := strings.TrimPrefix(url, "https:")

			// The registry line.
			if key == c.registryKey() && value == url {
				existingRegistries[url] = struct{}{}
			}

			// The always-auth line.
			if key == fmt.Sprintf("%s:always-auth", registry) {
				// Make sure it's set to true.
				newLine = fmt.Sprintf("%s:always-auth=true", registry)
				existingAuth[url] = struct{}{}
			}

			// The email line.
			if key == fmt.Sprintf("%s:email", registry) {
				existingEmails[url] = struct{}{}
			}

			// The _authToken line. This is the essential line to set the credential.
			if key == fmt.Sprintf("%s:_authToken", registry) {
				existingCreds[url] = struct{}{}
				authString := fmt.Sprintf("%s:%s", user, pwd)
				encodedAuth := base64.StdEncoding.EncodeToString([]byte(authString))
				newLine = fmt.Sprintf("%s:_authToken=%s", registry, encodedAuth)
			}
		}

		lines = append(lines, newLine)
	}

	if err := scanner.Err(); err != nil {
		panic(err) // Really shouldn't happen.
	}

	// Add missing registries.
	for _, repo := range repos {
		url := normalizeRepoURL(repo)
		registry := strings.TrimPrefix(url, "https:")

		if _, ok := existingRegistries[url]; !ok {
			lines = append(lines, fmt.Sprintf("%s=%s", c.registryKey(), url))
		}

		if _, ok := existingAuth[url]; !ok {
			lines = append(lines, fmt.Sprintf("%s:always-auth=true", registry))
		}

		if _, ok := existingEmails[url]; !ok {
			lines = append(lines, fmt.Sprintf("%s:email=not.valid@email.com", registry))
		}

		if _, ok := existingCreds[url]; !ok {
			authString := fmt.Sprintf("%s:%s", user, pwd)
			encodedAuth := base64.StdEncoding.EncodeToString([]byte(authString))
			lines = append(lines, fmt.Sprintf("%s:_authToken=%s", registry, encodedAuth))
		}
	}

	c.content = bytes.NewBufferString(strings.Join(lines, "\n") + "\n")
}

func (c *Config) registryKey() string {
	if c.scope == "" {
		return "registry"
	}
	return fmt.Sprintf("@%s:registry", c.scope)
}

func normalizeRepoURL(repoURL string) string {
	if !strings.HasPrefix(repoURL, "https://") {
		repoURL = "https://" + repoURL
	}
	if !strings.HasSuffix(repoURL, "/") {
		repoURL += "/"
	}
	return repoURL
}
