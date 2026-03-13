# GitHub MCP

A tool for interacting with GitHub API through MCP.

## Features

- List repositories and get repository details
- List and manage pull requests
- Create and comment on pull requests
- List, view, and manage issues
- Get file content from GitHub repositories

## Installation

There are several ways to install the GitHub Tool:

### Option 1: Download from GitHub Releases

1. Visit the [GitHub Releases](https://github.com/nguyenvanduocit/github-mcp/releases) page
2. Download the binary for your platform:
   - `github-mcp_linux_amd64` for Linux
   - `github-mcp_darwin_amd64` for macOS
   - `github-mcp_windows_amd64.exe` for Windows
3. Make the binary executable (Linux/macOS):
   ```bash
   chmod +x github-mcp_*
   ```
4. Move it to your PATH (Linux/macOS):
   ```bash
   sudo mv github-mcp_* /usr/local/bin/github-mcp
   ```

### Option 2: Go install
```
go install github.com/nguyenvanduocit/github-mcp
```

## Config

### Environment

1. Set up environment variables in `.env` file:
```
GITHUB_TOKEN=your_github_token
```

### Claude, cursor
```
{
  "mcpServers": {
    "github": {
      "command": "/path-to/github-mcp",
      "args": ["-env", "path-to-env-file"]
    }
  }
}
```

## CLI Usage

In addition to the MCP server, `github-mcp` ships a standalone CLI binary (`github-cli`) for direct terminal use — no MCP client needed.

### Installation

```bash
just install-cli
# or
go install github.com/nguyenvanduocit/github-mcp/cmd/cli@latest
```

### Quick Start

```bash
export GITHUB_TOKEN=your-github-token
# or
github-cli --env .env <command> [flags]
```

### Commands

| Command | Description |
|---------|-------------|
| `list-repos` | List repositories for a user/org |
| `get-repo` | Get repository details |
| `list-prs` | List pull requests |
| `get-pr` | Get pull request details |
| `create-pr` | Create a pull request |
| `create-pr-comment` | Comment on a pull request |
| `get-file` | Get file content from a repo |
| `list-issues` | List issues |
| `get-issue` | Get issue details |
| `comment-issue` | Comment on an issue |
| `issue-action` | Close or reopen an issue |
| `approve-pr` | Approve a pull request |

### Examples

```bash
# List repos
github-cli list-repos --owner myorg --type public

# List open PRs
github-cli list-prs --owner myorg --repo myrepo --state open

# Create a PR
github-cli create-pr --owner myorg --repo myrepo \
  --title "My PR" --head feature-branch --base main

# Get file content
github-cli get-file --owner myorg --repo myrepo --path src/main.go

# JSON output
github-cli list-issues --owner myorg --repo myrepo --output json | jq '.[].title'
```

### Flags

Every command accepts:
- `--env string` — Path to `.env` file
- `--output string` — Output format: `text` (default) or `json`

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
