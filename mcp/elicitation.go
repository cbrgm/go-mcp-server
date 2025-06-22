package mcp

import "context"

// ElicitationRequest represents a request from a server to gather additional information from the user.
//
// This allows servers to request structured data from users with JSON schemas to validate responses.
// Elicitation enables interactive workflows by allowing user input requests to occur nested inside
// other MCP server features.
type ElicitationRequest struct {
	// Prompt is the human-readable request for information.
	Prompt string `json:"prompt"`

	// Schema defines the expected structure of the user's response using JSON Schema.
	Schema map[string]any `json:"schema,omitempty"`

	// Meta contains implementation-specific metadata.
	Meta map[string]any `json:"_meta,omitempty"`
}

// ElicitationResponse contains the user's response to an elicitation request.
type ElicitationResponse struct {
	// Data contains the user's structured response.
	Data map[string]any `json:"data"`

	// Meta contains implementation-specific metadata.
	Meta map[string]any `json:"_meta,omitempty"`
}

// ElicitationHandler defines the interface for handling elicitation requests.
//
// This interface allows servers to request additional information from users
// during interactions, enabling interactive workflows while maintaining
// client control over user interactions and data sharing.
type ElicitationHandler interface {
	// HandleElicitation processes an elicitation request and returns the user's response.
	// The client maintains control over how the request is presented to the user
	// and how the response is collected.
	HandleElicitation(ctx context.Context, req ElicitationRequest) (ElicitationResponse, error)
}
