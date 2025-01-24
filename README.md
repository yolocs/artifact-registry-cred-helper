# artifact-registry-cred-helper


This CLI tool simplifies authentication with Google Cloud Artifact Registry. It manages your credentials stored in a .netrc file, allowing you to easily switch between different Artifact Registry repositories and authentication methods (OAuth 2.0 tokens and service account JSON keys).

## Installation

This project is assumed to be installed via go install. Ensure you have Go installed and configured correctly. Then, navigate to the root directory of this project and run:

```
go install github.com/yolocs/artifact-registry-cred-helper/cmd@latest
```

This will install the artifact-registry-cred-helper command in your $GOPATH/bin directory (or your system's PATH if configured differently). You may need to add this directory to your PATH environment variable.

## Usage

The tool provides two main commands: `set-netrc` and `get`.

### `set-netrc`

This command updates your .netrc file with credentials for specified Artifact Registry hosts. It supports both OAuth 2.0 tokens and service account JSON keys.

Syntax:

```
artifact-registry-cred-helper set-netrc [options]
```

Options:

* `--hosts <host1>,<host2>,...`: (Required unless --refresh is used) A comma-separated list of Artifact Registry hostnames (e.g., us-go.pkg.dev, eu-python.pkg.dev). Hostnames must end with .pkg.dev.
* `--json-key <path>`: The path to your service account JSON key file. If omitted, the tool uses an OAuth 2.0 token.
* `--netrc <path>`: The path to your .netrc file. Defaults to the standard system location.
* `--refresh`: Refreshes the credentials for existing Artifact Registry entries in your .netrc file. This option ignores the --hosts and --json-key flags. It only works with OAuth2.
* `--append`: Appends new entries to the .netrc file instead of overwriting existing Artifact Registry entries.

Examples:

Using OAuth 2.0 token:

```
artifact-registry-cred-helper set-netrc --hosts=us-go.pkg.dev,eu-python.pkg.dev
```

Using a service account JSON key:

```
artifact-registry-cred-helper set-netrc --hosts=us-central1.pkg.dev --json-key /path/to/your/key.json
```

Refreshing existing credentials:

```
artifact-registry-cred-helper set-netrc --refresh
```

Appending new hosts:

```
artifact-registry-cred-helper set-netrc --hosts=us-go.pkg.dev --append
```

### `get`

This command retrieves the credentials for a given host from your .netrc file. It's designed to be compatible with the Credential Helper Spec.

Syntax:

```
artifact-registry-cred-helper get [options]
```

Options:

* `--hosts <host1>,<host2>,...`: A comma-separated list of Artifact Registry hostnames. You can only use this OR stdin.
* **stdin**: A JSON payload with a "uri" field specifying the host. This is the preferred method for compatibility with the Credential Helper Spec.

Examples:

Using the --hosts flag:

```
artifact-registry-cred-helper get --hosts=us-go.pkg.dev
```

Using stdin (Credential Helper Spec compliant):

```
echo '{"uri": "us-go.pkg.dev"}' | artifact-registry-cred-helper get
```

The output will be a JSON object containing the authorization header to be used when interacting with the specified Artifact Registry host.

## Environment Variables

The following environment variables can be used to override command-line options:

* `AR_CRED_HELPER_HOSTS`: Overrides the `--hosts` flag.
* `AR_CRED_HELPER_JSON_KE`: Overrides the `--json-key` flag.
* `AR_CRED_HELPER_NETRC`: Overrides the `--netrc` flag.
* `AR_CRED_HELPER_REFRESH_ONLY`: Overrides the `--refresh` flag.
* `AR_CRED_HELPER_APPEND_ONLY`: Overrides the `--append` flag.
