# artifact-registry-cred-helper

This tool simplifies setting and managing credentials for accessing Google
Artifact Registry from various build tools and package managers, including
Maven, Python (pip), Go, and APT. Instead of relying on more complex solutions
like keyrings or individual tool-specific plugins (which introduce additional
dependencies), this tool uses the simple and widely-supported basic
authentication mechanism. This minimizes dependencies and simplifies
troubleshooting.

## Installation

This tool is a single binary. Download it from [insert release link here]. For
example, you might download `artifact-registry-cred-helper_linux_amd64`. Place
the binary in a directory included in your `PATH` environment variable.

Or:

```sh
go install github.com/yolocs/artifact-registry-cred-helper/cmd/artifact-registry-cred-helper@latest
```

## How it Works

`artifact-registry-cred-helper` modifies your build tool or package manager's
configuration file to include the necessary authentication credentials for your
Artifact Registry repositories. It uses a straightforward approach: It adds or
updates the relevant credential sections within your configuration file (e.g.,
`<server>` entries in Maven's `settings.xml`). This keeps your credentials
separate from your source code, improving security.

The tool currently supports (default credential files):

* **Maven:** Modifies `~/.m2/settings.xml`
* **Python (pip):**  Modifies `~/.netrc`
* **Go:** Modifies `~/.netrc`
* **APT:** Modifies `/etc/apt/auth.conf.d/artifact-registry.conf`

The tool supports two authentication methods supported by Artifact Registry:

* **OAuth2 Access Token:** **RECOMMENDED**. Suitable for short-lived tokens.
* **JSON Key (base64 encoded):** For long-term access.

## Usage

The tool itself provides detailed documentation.

```sh
artifact-registry-cred-helper -h
```

## For Local Developement

Always prefer short-lived tokens for authentication but it causes trouble for
local dev loop. The tool provides an option to refresh the credential in the
background until your `gcloud` is logged out. For example:

```sh
artifact-registry-cred-helper set-netrc \ 
  --repo-urls=us-go.pkg.dev/my-project/repo1,us-python.pkg.dev/my-project/repo2 \
  --background-refresh-interval=5m &
```

This command would run the tool in the background and refresh the credential
every 5 minutes in the `.netrc` file.

## For CI/CD

Coming soon.
