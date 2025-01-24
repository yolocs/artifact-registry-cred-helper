// Package netrc provides functions to modify an netrc file.
package netrc

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	// "github.com/GoogleCloudPlatform/artifact-registry-go-tools/pkg/auth"
)

var (
	catchAllPattern = regexp.MustCompile("\nmachine (.*.pkg.dev)\nlogin (oauth2accesstoken|_json_key_base64)\npassword (.*)\n")
	tokenPattern    = regexp.MustCompile("\nmachine (.*.pkg.dev)\nlogin oauth2accesstoken\npassword (.*)\n")
)

func tokenFormat(host, token string) string {
	return fmt.Sprintf(`
machine %s
login oauth2accesstoken
password %s
`, host, token)
}

func jsonKeyFormat(host, base64Key string) string {
	return fmt.Sprintf(`
machine %s
login _json_key_base64
password %s
`, host, base64Key)
}

type NetRC struct {
	path    string
	content string
}

func Open(netrcPath string) (*NetRC, error) {
	if netrcPath == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot open .netrc file: %w", err)
		}
		netrcPath = h
	}

	if !strings.HasSuffix(netrcPath, ".netrc") {
		netrcPath = path.Join(netrcPath, ".netrc")
	}

	if _, err := os.Stat(path.Dir(netrcPath)); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf(".netrc directory does not exist: %w", err)
		}
		return nil, fmt.Errorf("failed to load .netrc directory: %w", err)
	}

	data, err := os.ReadFile(netrcPath)
	if os.IsNotExist(err) {
		return &NetRC{path: netrcPath, content: ""}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot load .netrc file: %v", err)
	}
	return &NetRC{path: netrcPath, content: string(data)}, nil
}

func (n *NetRC) SetToken(hosts []string, token string, append bool) {
	n.update(hosts, tokenFormat, token, append)
}

func (n *NetRC) SetJSONKey(hosts []string, base64Key string, append bool) {
	n.update(hosts, jsonKeyFormat, base64Key, append)
}

func (n *NetRC) Refresh(token string) {
	n.content = tokenPattern.ReplaceAllString(n.content, "\nmachine $1\nlogin oauth2accesstoken\npassword "+token+"\n")
}

func (n *NetRC) Close() error {
	f, err := os.OpenFile(n.path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		return fmt.Errorf("failed to open .netrc: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(n.content); err != nil {
		return fmt.Errorf("failed to save .netrc: %w", err)
	}
	return nil
}

func (n *NetRC) update(hosts []string, formatter func(string, string) string, pwd string, append bool) {
	content := n.content
	if !append {
		// First clean up existing credentials.
		content = catchAllPattern.ReplaceAllString(n.content, "")
	}
	// Add new credentials.
	for _, h := range hosts {
		content = content + formatter(h, pwd)
	}
	n.content = content
}
