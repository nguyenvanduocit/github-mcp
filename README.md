# GitHub Tool

A tool for interacting with GitHub API through MCP.

## Features

- List repositories
- Get repository details
- List pull requests
- Get pull request details
- Create and comment on pull requests
- List, view, and manage issues
- Get file content from GitHub repositories

## Setup

1. Clone the repository
2. Set up environment variables in `.env` file:
   ```
   GITHUB_TOKEN=your_github_token
   ```
3. Build and run the tool

## Usage

Run the tool in SSE mode:
```
just dev
```

Or build and install:
```
just build
just install
```
