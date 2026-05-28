# Terraform Developer Docs MCP Server

A high-performance Model Context Protocol (MCP) server that provides local AI agents with fast, secure, and thread-safe read-only access to HashiCorp's official Terraform provider developer documentation. 

By cloning and caching the unified documentation repository (`web-unified-docs`), this server allows agents to list, read, and perform fast native searches across all documentation versions for building, testing, logging, and multiplexing Terraform providers.

---

## Capabilities & Toolset

The server registers three primary JSON-RPC tools. Every tool dynamically validates providers and versions, and falls back to the **latest released version** (highest semantic version folder) if `version` is omitted.

### 1. `list_docs`
Browse the directory structure and file trees of a developer library.
* **Arguments**:
  * `provider` (string, required): The target developer library key or alias.
  * `path` (string, optional): The relative subpath to list (e.g. `docs` or `" "` for root).
  * `version` (string, optional): Specific version subdirectory (e.g., `v1.18.x`, `v2.0.0`). If omitted, defaults to the latest stable release.
* **Response**: A newline-separated list of files and subdirectories (directories are suffixed with `/` for clarity).

### 2. `read_doc`
Fetch the complete markdown or text contents of a specific documentation file.
* **Arguments**:
  * `provider` (string, required): The target developer library key or alias.
  * `path` (string, required): Path to the documentation file relative to the version directory (e.g., `docs/index.md`).
  * `version` (string, optional): Specific version subdirectory (e.g., `v1.18.x`, `v2.0.0`). If omitted, defaults to the latest stable release.
* **Response**: The raw text content of the markdown/text file.

### 3. `search_docs`
Perform ultra-fast, case-insensitive literal string searches across all documentation files using native multi-threaded `git grep`.
* **Arguments**:
  * `provider` (string, required): The target developer library key or alias.
  * `query` (string, required): The literal text snippet or keyword to search for.
  * `version` (string, optional): Specific version subdirectory (e.g., `v1.18.x`, `v2.0.0`). If omitted, defaults to the latest stable release.
* **Response**: A formatted list of matching files, including exact matching line numbers and source snippets, capped at 100 results to optimize LLM context usage.

---

## Supported Developer Libraries

The server maps descriptive key names and backward-compatible aliases to their respective subdirectories within the unified documentation mono-repo:

| Verbose Key (Preferred) | Alias (Backward Compatibility) | Target Path in Unified Docs | Description |
| :--- | :--- | :--- | :--- |
| `"plugin-framework"` | `"framework"` | `content/terraform-plugin-framework` | Modern Go Plugin Framework |
| `"plugin-sdk-v2"` | `"sdk"`, `"sdkv2"` | `content/terraform-plugin-sdk` | Legacy SDKv2 |
| `"terraform-core"` | `"core"` | `content/terraform` | Terraform Core / Integration Protocol |
| `"plugin-testing"` | `"testing"` | `content/terraform-plugin-testing` | Provider Testing Framework |
| `"plugin-go"` | *None* | `content/terraform-plugin-go` | Lower-level Go Plugin Bindings |
| `"plugin-log"` | *None* | `content/terraform-plugin-log` | Framework Logging Utilities |
| `"plugin-mux"` | *None* | `content/terraform-plugin-mux` | SDKv2 and Framework Multiplexer |

---

## Architectural & Safety Features

* **Single-Repository Cache**: Instead of duplicate cloning, the server clones `web-unified-docs` exactly once to the cache directory on startup.
  * The cache location defaults to `.docs_cache`.
  * You can customize the location using the `TERRAFORM_DOCS_CACHE_DIR` environment variable.
  * If the server is running on a **Read-Only (RO) filesystem**, it automatically detects that the directory is not writeable and falls back to a safe subdirectory under the system's temporary directory (`os.TempDir()`), typically `/tmp/terraform-provider-docs-cache` or similar.
* **RWMutex Thread Safety**: Uses a `sync.RWMutex` to ensure thread-safe operation. Multiple client queries can read or search in parallel, while background repository synchronization operations are safely isolated.
* **Local Sandboxing**: Operates strictly read-only and over stdio transport. No TCP/UDP network ports are opened, ensuring complete isolation and protection from cross-origin exploits or port clashes.

---

## Installation

### Standard Go Install (Recommended)
You can install the MCP server globally without needing to clone this repository manually. Run the following command:

```bash
go install github.com/jackofallops/mcp-terraform-provider-dev-docs/cmd/terraform-docs-mcp@latest
```

This compiles and installs the `terraform-docs-mcp` executable into your standard Go binary path (typically `~/go/bin/terraform-docs-mcp` on macOS/Linux, or `%USERPROFILE%\go\bin\terraform-docs-mcp.exe` on Windows).

### Local Compilation
If you have cloned the repository locally, you can build it using:

```bash
go build -o terraform-docs-mcp cmd/terraform-docs-mcp/main.go
```

---

## IDE Configuration

Configure your AI client to use the server by adding it to your client config. Below are standard configuration examples utilizing the globally installed binary. (Replace `~/go/bin/` with the absolute path to your home directory, e.g. `/Users/yourusername/go/bin/terraform-docs-mcp`).

### 1. Claude Desktop
Add the following to your `~/.claude/claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "terraform_docs": {
      "command": "/Users/YOUR_USERNAME/go/bin/terraform-docs-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

### 2. Cursor
Add to your global Cursor settings or `.cursor/mcp.json`:
```json
{
  "mcpServers": {
    "terraform_docs": {
      "command": "/Users/YOUR_USERNAME/go/bin/terraform-docs-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

### 3. Antigravity
Configure in your global `~/.antigravity/mcp.json` or workspace settings:
```json
{
  "mcpServers": {
    "terraform_docs": {
      "command": "/Users/YOUR_USERNAME/go/bin/terraform-docs-mcp",
      "args": [],
      "env": {}
    }
  }
}
```
