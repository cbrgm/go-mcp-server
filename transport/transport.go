// Package transport provides MCP transport layer implementations.
//
// This package defines the Transport interface and provides implementations
// for different transport mechanisms supported by the MCP specification:
//   - Stdio transport for process-based communication
//   - HTTP transport for network-based communication
//
// All transports use JSON-RPC 2.0 for message exchange and support the
// full MCP protocol including initialization, requests, and responses.
package transport

import (
	"context"

	"github.com/cbrgm/go-mcp-server/server"
)

// Transport defines the interface for MCP transport mechanisms.
//
// Transport implementations handle the low-level communication details
// while delegating MCP protocol logic to the server. Each transport
// is responsible for message framing, encoding/decoding, and error handling.
type Transport interface {
	// Start begins listening for requests on this transport.
	// It blocks until the context is cancelled or an error occurs.
	Start(ctx context.Context, server *server.Server) error

	// Stop gracefully shuts down the transport.
	// It should stop accepting new connections and wait for existing
	// requests to complete before returning.
	Stop() error
}
