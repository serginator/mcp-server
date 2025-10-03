# MCP Integration Server

Curated MCP server that explores GitHub, Jira, and Notion tooling while staying comfortably under the 80-tool guidance for editor integrations such as Cursor.

## Tooling Overview

### GitHub Tools (21)

- `github_get_pull_request` – Get details of a specific pull request
- `github_get_pull_request_diff` – Fetch a pull request diff in unified format
- `github_create_issue` – Create a new issue in a repository
- `github_create_pull_request` – Create a new pull request
- `github_get_issue` – Get details of a specific issue
- `github_list_branches` – List all branches in a repository
- `github_list_commits` – List commits in a repository
- `github_search_repositories` – Search for repositories
- `github_search_issues` – Search for issues across repositories
- `github_get_workflows` – List GitHub Actions workflows
- `github_run_workflow` – Trigger a workflow run by file name
- `github_add_comment` – Add a comment to an issue or pull request
- `github_get_comments` – Retrieve comments from an issue or pull request
- `github_assign_copilot` – Assign users to an issue or pull request
- `github_create_branch` – Create a new branch from a commit SHA
- `github_create_repository` – Create a repository in the authenticated account
- `github_get_commit` – Get details for a specific commit
- `github_get_release_by_tag` – Retrieve release information by tag name
- `github_get_tag` – Look up tag metadata and commit
- `github_search_code` – Search for code results across repositories
- `github_search_pull_requests` – Search for pull requests

### Jira Tools (3)

- `jira_get_ticket` – Get details of a Jira ticket
- `jira_search_tickets` – Search for Jira tickets using JQL
- `jira_create_ticket` – Create a new Jira ticket as a task

### Notion Tools (7)

- `notion_search_pages` – Search for Notion pages by title
- `notion_get_page` – Get a Notion page by URL
- `notion_get_database` – Get a Notion database by ID
- `notion_create_page` – Create a new Notion page under a parent page
- `notion_create_database` – Create a new Notion database under a parent page
- `notion_update_page` – Update metadata for an existing page
- `notion_update_database` – Update the title for an existing database

## Configuration

Create a `config.yml` file with your API tokens:

```yaml
notion_token: "your_notion_integration_token"
github_token: "your_github_personal_access_token"
jira_token: "your_jira_api_token"
jira_url: "https://your-domain.atlassian.net/"
```

## Running with Docker

1. Build and start the server:

```bash
docker-compose up --build
```

2. The server will listen on stdin/stdout using the MCP protocol.

## Running Locally

1. Install dependencies:

```bash
make dependencies
```

2. Build:

```bash
make build
```

## Testing

Use the provided test script:

```bash
./test-mcp.sh
```

Or test manually using JSON-RPC over stdin:

### Initialize the server:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}' | ./mcp-server
```

### List available tools:

```bash
echo -e '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}\n{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./mcp-server
```

### Call a tool:

```bash
echo -e '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}\n{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"github_search_repositories","arguments":{"query":"golang mcp"}}}' | ./mcp-server
```

## Integration with AI Models

This MCP server can be integrated with:

- Cursor
- Claude Desktop
- ChatGPT with MCP support
- Any MCP-compatible AI client

Configure your AI client to connect to this server using the MCP protocol over stdio. Normally is just configuring the `mcp.json` file with the following content:

```json
{
  "mcpServers": {
    "mcp-local-server": {
      "command": "<your-path>/mcp-server/run-docker-mcp.sh",
      "args": [],
      "env": {}
    }
  }
}
```

## API Token Setup

### GitHub

1. Go to GitHub Settings > Developer settings > Personal access tokens
2. Generate a token with appropriate permissions (repo, workflow, etc.)

### Jira

1. Go to Jira Settings > Security > API tokens
2. Create an API token for your account

### Notion

1. Go to https://www.notion.so/my-integrations
2. Create a new integration
3. Copy the integration token

## Architecture

The server follows a modular architecture:

- `main.go` - Entry point and configuration loading
- `server/` - MCP protocol implementation
- `tools/` - Tool interface definitions
- `github/`, `jira/`, `notion/` - Service implementations

## Security Notes

- Keep your API tokens secure and never commit them to version control
- Use environment variables or secure configuration management in production
- Consider rate limiting and access controls for production deployments

## License

MIT License
