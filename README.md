# Terraform Docs MCP Server

A lightweight Model Context Protocol (MCP) server that provides AI agents with read-only access to Terraform provider documentation. It securely clones and caches specific git repositories, enabling intelligent navigation, searching, and reading of markdown-based API references and guides.

## Capabilities

- **List**: Browse directory structures and file trees within a provider's documentation.
- **Read**: Fetch the raw content of any `.md` file by path.
- **Search**: Perform full-text searches across documentation, returning relevant snippets.
- **Version Control**: Pin to specific git tags or branches (e.g., `v1.6.0`, `main`) for reproducible results.

Supported providers: `framework`, `sdk`, `core`.

## IDE Configuration Examples

Configure your AI agent to use this server by adding the following JSON configuration to your IDE's MCP settings file. Replace `/path/to/terraform-docs-mcp` with the actual executable path on your system.

### Claude Desktop
Add to `~/.claude/claude_desktop_config.json`:
Add to `~/.claude/claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "terraform_docs": {
      "command": "/path/to/terraform-docs-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

### Cursor

Add to your project's .cursor/mcp.json (or global settings):

```json
{
  "mcpServers": {
    "terraform_docs": {
      "command": "/path/to/terraform-docs-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

### Antigravity

Configure in your IDE's MCP client settings or ~/.antigravity/mcp.json:

```json
{
  "mcpServers": {
    "terraform_docs": {
      "command": "/path/to/terraform-docs-mcp",
      "args": [],
      "env": {}
    }
  }
}
```


### Usage Notes

* The server caches cloned repositories in a configurable directory to avoid repeated network requests.  
* All operations are read-only and safe for AI agents.  
* If no version is specified, the server uses the currently checked-out state of the cached repository.  
* Ensure your IDE's MCP client supports standard input/output (stdio) transport.  
