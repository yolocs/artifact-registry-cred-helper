// Package netrc provides functions to modify an netrc file.
package netrc

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
			return nil, fmt.Errorf("cannot find HOME dir: %w", err)
		}
		netrcPath = h
	}

	data, err := os.ReadFile(netrcPath)
	if os.IsNotExist(err) {
		return &NetRC{path: netrcPath, content: ""}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot load file %q: %v", netrcPath, err)
	}
	return &NetRC{path: netrcPath, content: string(data)}, nil
}

func (n *NetRC) SetToken(hosts []string, token string) {
	n.update(hosts, tokenFormat, token, false)
}

func (n *NetRC) SetJSONKey(hosts []string, base64Key string) {
	n.update(hosts, jsonKeyFormat, base64Key, false)
}

func (n *NetRC) Refresh(token string) {
	n.content = tokenPattern.ReplaceAllString(n.content, "\nmachine $1\nlogin oauth2accesstoken\npassword "+token+"\n")
}

func (n *NetRC) Close() error {
	// Make sure dir exists.
	if err := os.MkdirAll(filepath.Dir(n.path), 0755); err != nil {
		return fmt.Errorf("failed to create dir for %q: %w", n.path, err)
	}

	f, err := os.OpenFile(n.path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		return fmt.Errorf("failed to open %q: %w", n.path, err)
	}
	defer f.Close()

	if _, err := f.WriteString(n.content); err != nil {
		return fmt.Errorf("failed to save %q: %w", n.path, err)
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
