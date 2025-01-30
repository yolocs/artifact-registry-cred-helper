package apt

import (
	"fmt"
	"path/filepath"

	"github.com/yolocs/artifact-registry-cred-helper/pkg/netrc"
)

const (
	configDir = "/etc/apt/apt.conf.d"
)

// AuthConfig represents a apt auth config file.
// It uses the same format as netrc. So to reuse its implementation, here we
// set all artifact registry entries in a single file.
type AuthConfig struct {
	config *netrc.NetRC
}

func Open(configName string) (*AuthConfig, error) {
	if configName == "" {
		configName = "artifact-registry.conf"
	}

	configPath := filepath.Join(configDir, configName)
	config, err := netrc.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open apt auth config file (as netrc) at %q: %w", configPath, err)
	}

	return &AuthConfig{config: config}, nil
}

func (c *AuthConfig) Close() error {
	return c.config.Close()
}

func (c *AuthConfig) SetToken(hosts []string, token string) {
	c.config.SetToken(hosts, token, false)
}

func (c *AuthConfig) SetJSONKey(hosts []string, base64Key string) {
	c.config.SetJSONKey(hosts, base64Key, false)
}
