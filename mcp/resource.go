package mcp

// Resource represents a piece of data or content that can be read by the client.
//
// Resources provide contextual information that can be used by LLMs. They are
// identified by URIs and can contain various types of content such as text,
// structured data, or references to external systems.
type Resource struct {
	// URI is the unique identifier for the resource.
	// It should follow a consistent scheme (e.g., "file://", "db://").
	URI string `json:"uri"`

	// Name is a human-readable name for the resource.
	Name string `json:"name"`

	// Title is a human-friendly display name for the resource.
	// TODO: Add back when upgrading to newer MCP spec
	// Title string `json:"title,omitempty"`

	// Meta contains implementation-specific metadata.
	// TODO: Add back when upgrading to newer MCP spec
	// Meta map[string]any `json:"_meta,omitempty"`
}

// ResourceContent contains the actual content of a resource.
//
// When a resource is read, the server returns the content along with the URI
// for identification. Content can be text, structured data, or other formats.
type ResourceContent struct {
	// URI identifies which resource this content belongs to.
	URI string `json:"uri"`

	// Text contains the textual content of the resource.
	Text string `json:"text"`
}

// ResourceResponse is the response to a resource read request.
//
// A single resource request can return multiple content items, for example
// when reading a directory that contains multiple files or entries.
type ResourceResponse struct {
	// Contents contains the resource content items.
	Contents []ResourceContent `json:"contents"`
}

// ResourceTemplate represents a parameterized resource using URI templates.
//
// Resource templates allow servers to expose dynamic resources that can be
// instantiated with different parameters. They use URI templates to define
// the structure of the resource URIs.
type ResourceTemplate struct {
	// URITemplate is the URI template for the resource (e.g., "file:///{path}").
	URITemplate string `json:"uriTemplate"`

	// Name is a human-readable name for the resource template.
	Name string `json:"name"`

	// Description explains what the resource template provides.
	Description string `json:"description,omitempty"`

	// MimeType indicates the MIME type of resources created from this template.
	MimeType string `json:"mimeType,omitempty"`

	// TODO: Add back when upgrading to newer MCP spec
	// Title is a human-friendly display name for the resource template.
	// Title string `json:"title,omitempty"`

	// TODO: Add back when upgrading to newer MCP spec
	// Meta contains implementation-specific metadata.
	// Meta map[string]any `json:"_meta,omitempty"`
}

// ResourceParams contains the parameters for reading a resource.
type ResourceParams struct {
	// URI identifies the resource to read.
	URI string `json:"uri"`
}
