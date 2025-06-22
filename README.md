# Go MCP Server

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server implementation in Go that provides tea information through tools, resources, and prompts.

**Note**: This project is for learning purposes only. I built this MCP server from scratch to understand the Model Context Protocol specification. For production use, consider using [mcp-go](https://github.com/mark3labs/mcp-go) or wait for the official MCP Go SDK to be released (expected end of June 2025).

This MCP server, especially the `mcp` package, was written with the help of [Claude Code](https://claude.ai/code) using the latest version of the [MCP specification](https://modelcontextprotocol.io/llms-full.txt) (2025-06-21).

## Features

- **MCP 2025-03-26 Specification Compliant**
- **Multiple Transports**: `stdio` (default), `http` with SSE
- **Tea Collection**: 8 premium teas (Green, Black, Oolong, White)
- **Full MCP Capabilities**: Tools, Resources, and Prompts

## Quick Start

```bash
# Build the binary
go build ./cmd/go-mcp-server

# Run with stdio transport (default)
./go-mcp-server

# Run with HTTP transport
./go-mcp-server -transport http -port 8080

# Test with MCP Inspector
echo '{"jsonrpc":"2.0","method":"initialize","id":1}' | ./go-mcp-server
```

## Usage

The server binary accepts several command line arguments to configure its behavior:

```bash
./go-mcp-server [options]
```

### Command Line Arguments

| Argument | Type | Default | Description |
|----------|------|---------|-------------|
| `-transport` | string | `stdio` | Transport protocol to use (`stdio` or `http`) |
| `-port` | int | `8080` | HTTP server port (only used with `-transport http`) |
| `-request-timeout` | duration | `30s` | Maximum time to wait for request processing |
| `-shutdown-timeout` | duration | `10s` | Maximum time to wait for graceful shutdown |
| `-log-level` | string | `info` | Log level (`debug`, `info`, `warn`, `error`) |
| `-log-json` | bool | `false` | Output logs in JSON format |
| `-server-name` | string | `MCP Server` | Server name returned in initialization |
| `-server-version` | string | `1.0.0` | Server version returned in initialization |

### Examples

```bash
# Default stdio transport with debug logging
./go-mcp-server -log-level debug

# HTTP transport on custom port with JSON logs
./go-mcp-server -transport http -port 9000 -log-json

# Custom timeouts for production use
./go-mcp-server -transport http -request-timeout 60s -shutdown-timeout 30s

# Custom server identification
./go-mcp-server -server-name "My Tea Server" -server-version "2.0.0"
```

## MCP Capabilities

### Tools
- `getTeaNames` - List all available teas
- `getTeaInfo` - Get detailed tea information and brewing instructions
- `getTeasByType` - Filter teas by type (Green Tea, Black Tea, Oolong Tea, White Tea)

### Resources
- `menu://tea` - Complete tea collection with prices and details

### Prompts
- `tea_recommendation` - Personalized recommendations based on mood/preferences
- `brewing_guide` - Detailed brewing instructions for specific teas
- `tea_pairing` - Food pairing suggestions

## Tea Collection Example

Try these commands to explore the tea collection:

```bash
# List all available teas
echo '{"jsonrpc":"2.0","method":"tools/call","id":1,"params":{"name":"getTeaNames","arguments":{}}}' | ./go-mcp-server

# Get information about a specific tea
echo '{"jsonrpc":"2.0","method":"tools/call","id":2,"params":{"name":"getTeaInfo","arguments":{"name":"earl-grey"}}}' | ./go-mcp-server

# Get all oolong teas
echo '{"jsonrpc":"2.0","method":"tools/call","id":3,"params":{"name":"getTeasByType","arguments":{"type":"Oolong Tea"}}}' | ./go-mcp-server

# Read the complete tea menu resource
echo '{"jsonrpc":"2.0","method":"resources/read","id":4,"params":{"uri":"menu://tea"}}' | ./go-mcp-server

# Get a brewing guide for gyokuro
echo '{"jsonrpc":"2.0","method":"prompts/get","id":5,"params":{"name":"brewing_guide","arguments":{"tea_name":"gyokuro"}}}' | ./go-mcp-server
```

## Web UI

When using HTTP transport, a web status page is available at the root path (`/`) of the server. This page shows server information, active sessions, and available endpoints.

## MCP Client Configuration

### Claude Desktop / VS Code / Other MCP Clients

Add this configuration to your MCP client settings:

```json
{
  "mcpServers": {
    "tea": {
      "command": "podman",
      "args": ["run", "-i", "--rm", "ghcr.io/cbrgm/go-mcp-server:v1"]
    }
  }
}
```

For local development, you can also use:

```json
{
  "mcpServers": {
    "tea": {
      "command": "go",
      "args": ["run", "./cmd/go-mcp-server"],
      "cwd": "/path/to/go-mcp-server"
    }
  }
}
```

## Testing with MCP Inspector

1. Install: `npm install -g @modelcontextprotocol/inspector`
2. Start: `npx @modelcontextprotocol/inspector`
3. Connect with command: `go run ./cmd/go-mcp-server`

## License

Apache 2.0
