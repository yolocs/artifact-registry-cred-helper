package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/abcxyz/pkg/cli"
)

type GetCommand struct {
	baseCommand

	hosts []string
	// jsonKey string
}

func (c *GetCommand) Desc() string {
	return "Get the credential for the given host."
}

func (c *GetCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

There are two ways to specify the host(s) to get credential for:

1. Use flag --hosts.

2. Pass a request JSON payload to the stdin. 
   See schema at: https://github.com/EngFlow/credential-helper-spec/blob/main/schemas/get-credentials-request.schema.json.
   This makes the tool comformant to the credential-helper-spec.
`
}

func (c *GetCommand) Flags() *cli.FlagSet {
	set := c.NewFlagSet()
	sec := set.NewSection("OPTIONS")

	sec.StringSliceVar(&cli.StringSliceVar{
		Name:    "hosts",
		Usage:   "The hosts to get the credential for. Must have domain '*.pkg.dev'.",
		Target:  &c.hosts,
		EnvVar:  "AR_CRED_HELPER_HOSTS",
		Example: "us-go.pkg.dev,eu-python.pkg.dev,asia-maven.pkg.dev",
	})

	// sec.StringVar(&cli.StringVar{
	// 	Name:   "json-key",
	// 	Usage:  "The path to the JSON key of a service account used for authentication. Leave empty to use oauth token instead.",
	// 	Target: &c.jsonKey,
	// 	EnvVar: "AR_CRED_HELPER_JSON_KEY",
	// })

	return set
}

func (c *GetCommand) Run(ctx context.Context, args []string) error {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	var hosts []string
	if len(c.hosts) > 0 {
		hosts = c.hosts
	} else {
		b, err := io.ReadAll(c.Stdin())
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		h, err := decodeReq(b)
		if err != nil {
			return err
		}
		hosts = []string{h}
	}

	if err := validateHosts(hosts); err != nil {
		return err
	}

	token, err := c.getAuthToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	out, err := encodeResp(token)
	if err != nil {
		return err
	}

	if _, err := c.Stdout().Write(out); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}

func decodeReq(req []byte) (string, error) {
	v := map[string]string{}
	if err := json.Unmarshal(req, &v); err != nil {
		return "", fmt.Errorf("failed to parse request: %w", err)
	}
	return v["uri"], nil
}

func encodeResp(token string) ([]byte, error) {
	v := map[string]any{
		"headers": map[string][]string{
			"Authorization": {"Bearer " + token},
		},
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to encode response: %w", err)
	}
	return b, nil
}
