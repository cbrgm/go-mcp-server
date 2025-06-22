// Package mcp provides core Model Context Protocol types and interfaces.
//
// This package implements the Model Context Protocol (MCP) specification version 2025-03-26,
// enabling communication between LLM applications (hosts) and context providers (servers).
//
// The MCP follows a client-server architecture where:
//   - Hosts are LLM applications that initiate connections
//   - Clients maintain 1:1 connections with servers inside the host application
//   - Servers provide context, tools, and prompts to clients
//
// All communication uses JSON-RPC 2.0 over various transport mechanisms.
package mcp

import (
	"context"
)

const (
	// ProtocolVersion defines the MCP protocol version this implementation supports.
	ProtocolVersion = "2025-03-26"

	// JSONRPCVersion defines the JSON-RPC version used for all MCP communications.
	JSONRPCVersion = "2.0"
)

// Standard JSON-RPC 2.0 error codes as defined in the specification.
const (
	// ErrorCodeParseError indicates invalid JSON was received.
	ErrorCodeParseError = -32700

	// ErrorCodeInvalidRequest indicates the JSON sent is not a valid Request object.
	ErrorCodeInvalidRequest = -32600

	// ErrorCodeMethodNotFound indicates the method does not exist or is not available.
	ErrorCodeMethodNotFound = -32601

	// ErrorCodeInvalidParams indicates invalid method parameter(s).
	ErrorCodeInvalidParams = -32602

	// ErrorCodeInternalError indicates internal JSON-RPC error.
	ErrorCodeInternalError = -32603
)

// ServerInfo contains metadata about an MCP server implementation.
type ServerInfo struct {
	// Name is the human-readable name of the server.
	Name string `json:"name"`

	// Version is the version of the server implementation.
	Version string `json:"version"`
}

// InitializeResponse is sent by the server in response to an initialize request.
// It contains the server's protocol version, capabilities, and metadata.
type InitializeResponse struct {
	// ProtocolVersion is the MCP protocol version the server supports.
	ProtocolVersion string `json:"protocolVersion"`

	// Capabilities describes what the server can do (tools, resources, prompts).
	Capabilities map[string]any `json:"capabilities"`

	// ServerInfo contains metadata about the server.
	ServerInfo ServerInfo `json:"serverInfo"`
}

// Request represents a JSON-RPC 2.0 request message.
type Request struct {
	// JSONRPC must be exactly "2.0" to indicate JSON-RPC 2.0.
	JSONRPC string `json:"jsonrpc"`

	// Method is the name of the method to be invoked.
	Method string `json:"method"`

	// ID is the request identifier. Must be string, number, or null.
	// For MCP, ID MUST NOT be null per specification.
	ID any `json:"id"`

	// Params contains the parameter values to be used during method invocation.
	Params any `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response message.
type Response struct {
	// JSONRPC must be exactly "2.0" to indicate JSON-RPC 2.0.
	JSONRPC string `json:"jsonrpc"`

	// ID must match the ID of the request being responded to.
	ID any `json:"id"`

	// Result contains the result of the method invocation.
	// This field is required on success and must not exist if there was an error.
	Result any `json:"result,omitempty"`

	// Error contains error information if the method invocation failed.
	// This field is required on error and must not exist if successful.
	Error *ErrorResponse `json:"error,omitempty"`
}

// ErrorResponse represents a JSON-RPC 2.0 error object.
type ErrorResponse struct {
	// Code is a numeric error code indicating the type of error.
	Code int `json:"code"`

	// Message is a short description of the error.
	Message string `json:"message"`

	// Data provides additional information about the error.
	Data any `json:"data,omitempty"`
}

// Server defines the core MCP server interface.
//
// Server implementations handle the MCP initialization handshake and process
// incoming JSON-RPC requests according to the MCP specification.
type Server interface {
	// Initialize handles the MCP initialization handshake.
	// It returns the server's protocol version, capabilities, and metadata.
	Initialize(ctx context.Context) (*InitializeResponse, error)

	// HandleRequest processes a JSON-RPC request and sends the appropriate response.
	// The context may contain a ResponseSender for sending responses back to the client.
	HandleRequest(ctx context.Context, req Request) error
}

// ToolHandler defines the interface for handling MCP tool operations.
//
// Tools are functions that the server can execute on behalf of the client.
// They represent actions that can be taken in the external system.
type ToolHandler interface {
	// ListTools returns all available tools that can be called.
	// Tools should include their name, description, and input schema.
	ListTools(ctx context.Context) ([]Tool, error)

	// CallTool executes a tool with the given parameters.
	// It returns the result of the tool execution or an error if the call fails.
	CallTool(ctx context.Context, params ToolCallParams) (ToolResponse, error)
}

// ResourceHandler defines the interface for handling MCP resource operations.
//
// Resources represent data or content that can be read by the client.
// They provide contextual information that can be used by LLMs.
type ResourceHandler interface {
	// ListResources returns all available resources that can be read.
	// Resources should include their URI and human-readable name.
	ListResources(ctx context.Context) ([]Resource, error)

	// ReadResource reads the content of a specific resource identified by URI.
	// It returns the resource content or an error if the resource cannot be read.
	ReadResource(ctx context.Context, params ResourceParams) (ResourceResponse, error)

	// ListResourceTemplates returns all available resource templates.
	// Resource templates define parameterized resources using URI templates.
	ListResourceTemplates(ctx context.Context) ([]ResourceTemplate, error)
}

// PromptHandler defines the interface for handling MCP prompt operations.
//
// Prompts are templates that can be used to generate structured interactions
// with language models, helping to standardize common use cases.
type PromptHandler interface {
	// ListPrompts returns all available prompt templates.
	// Prompts should include their name, description, and expected arguments.
	ListPrompts(ctx context.Context) ([]Prompt, error)

	// GetPrompt generates a prompt with the given parameters.
	// It returns formatted messages ready for use with a language model.
	GetPrompt(ctx context.Context, params PromptParams) (PromptResponse, error)
}

// ResponseSender defines the interface for sending responses back to clients.
//
// ResponseSender abstracts the transport mechanism, allowing the same server
// logic to work with different transport layers (stdio, HTTP, etc.).
type ResponseSender interface {
	// SendResponse sends a successful JSON-RPC response.
	SendResponse(response Response) error

	// SendError sends a JSON-RPC error response with the specified error details.
	SendError(id any, code int, message string, data any) error
}

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// ResponseSenderKey is the context key for accessing the ResponseSender.
	ResponseSenderKey contextKey = "responseSender"

	// SessionIDKey is the context key for accessing the session identifier.
	SessionIDKey contextKey = "sessionID"
)
