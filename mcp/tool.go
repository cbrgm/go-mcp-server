package mcp

// Tool represents a function that can be called by the MCP client.
//
// Tools are executable functions that allow the client to perform actions
// in the external system that the MCP server represents. Each tool has a
// name, description, and JSON schema defining its input parameters.
type Tool struct {
	// Name is the unique identifier for the tool.
	Name string `json:"name"`

	// Title is a human-friendly display name for the tool.
	// TODO: Add back when upgrading to newer MCP spec
	// Title string `json:"title,omitempty"`

	// Description explains what the tool does and when to use it.
	Description string `json:"description"`

	// InputSchema defines the expected parameters using JSON Schema.
	InputSchema InputSchema `json:"inputSchema"`

	// Meta contains implementation-specific metadata.
	// TODO: Add back when upgrading to newer MCP spec
	// Meta map[string]any `json:"_meta,omitempty"`
}

// InputSchema defines the JSON Schema for tool input parameters.
//
// This follows the JSON Schema specification and describes what parameters
// the tool expects, their types, and which ones are required.
type InputSchema struct {
	// Type is typically "object" for tool parameters.
	Type string `json:"type"`

	// Properties defines the individual parameter schemas.
	Properties map[string]any `json:"properties,omitempty"`

	// Required lists the parameter names that must be provided.
	Required []string `json:"required,omitempty"`
}

// ToolCallParams contains the parameters for calling a tool.
type ToolCallParams struct {
	// Name is the name of the tool to call.
	Name string `json:"name"`

	// Arguments contains the parameters to pass to the tool.
	Arguments map[string]any `json:"arguments"`
}

// ToolResponse contains the result of a tool execution.
//
// The response contains content items that represent the output of the tool.
// Content can be text, images, or other media types supported by MCP.
type ToolResponse struct {
	// Content contains the output of the tool execution.
	Content []ContentItem `json:"content"`
}

// ContentItem represents a piece of content in a tool response.
//
// Content items can be text, images, or other media types. The type field
// indicates what kind of content this is, and additional fields provide
// the actual content data.
type ContentItem struct {
	// Type indicates the content type (e.g., "text", "image", "resource").
	Type string `json:"type"`

	// Text contains the text content when Type is "text".
	Text string `json:"text,omitempty"`

	// Data contains the raw data when Type is "image" or other binary content.
	Data string `json:"data,omitempty"`

	// MimeType specifies the MIME type for binary content.
	MimeType string `json:"mimeType,omitempty"`

	// Resource contains a reference to an MCP resource when Type is "resource".
	Resource *ResourceReference `json:"resource,omitempty"`
}

// ResourceReference represents a reference to an MCP resource in tool output.
//
// This allows tools to link to resources that provide additional context
// or data related to the tool's execution.
type ResourceReference struct {
	// URI identifies the resource being referenced.
	URI string `json:"uri"`

	// Type indicates the type of resource reference.
	Type string `json:"type,omitempty"`
}
