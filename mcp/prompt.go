package mcp

// Prompt represents a template for generating structured LLM interactions.
//
// Prompts help standardize common use cases by providing templates that can
// be customized with arguments. They generate messages ready for use with
// language models.
type Prompt struct {
	// Name is the unique identifier for the prompt.
	Name string `json:"name"`

	// Title is a human-friendly display name for the prompt.
	// TODO: Add back when upgrading to newer MCP spec
	// Title string `json:"title,omitempty"`

	// Description explains what the prompt does and when to use it.
	Description string `json:"description"`

	// Arguments defines the parameters this prompt accepts.
	Arguments []PromptArgument `json:"arguments,omitempty"`

	// Meta contains implementation-specific metadata.
	// TODO: Add back when upgrading to newer MCP spec
	// Meta map[string]any `json:"_meta,omitempty"`
}

// PromptArgument defines a parameter that can be passed to a prompt.
//
// Arguments allow prompts to be customized for different contexts while
// maintaining a consistent structure and behavior.
type PromptArgument struct {
	// Name is the parameter name.
	Name string `json:"name"`

	// Description explains what this argument is used for.
	Description string `json:"description"`

	// Required indicates whether this argument must be provided.
	Required bool `json:"required,omitempty"`
}

// PromptParams contains the parameters for generating a prompt.
type PromptParams struct {
	// Name is the name of the prompt to generate.
	Name string `json:"name"`

	// Arguments contains the values for the prompt parameters.
	Arguments map[string]any `json:"arguments,omitempty"`
}

// PromptResponse contains the generated prompt messages.
//
// The response contains a sequence of messages that form a conversation
// template ready for use with a language model.
type PromptResponse struct {
	// Messages contains the generated conversation messages.
	Messages []PromptMessage `json:"messages"`
}

// PromptMessage represents a single message in a generated prompt.
//
// Messages follow the standard conversation format with roles like "user",
// "assistant", or "system" and contain the actual message content.
type PromptMessage struct {
	// Role indicates who is speaking ("user", "assistant", "system").
	Role string `json:"role"`

	// Content contains the message content.
	Content MessageContent `json:"content"`
}

// MessageContent contains the actual content of a prompt message.
//
// Content can be text or other media types. The type field indicates
// what kind of content this is.
type MessageContent struct {
	// Type indicates the content type (typically "text").
	Type string `json:"type"`

	// Text contains the text content when Type is "text".
	Text string `json:"text"`
}
