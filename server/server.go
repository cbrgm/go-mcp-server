package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/cbrgm/go-mcp-server/mcp"
)

type Server struct {
	toolHandler     mcp.ToolHandler
	resourceHandler mcp.ResourceHandler
	promptHandler   mcp.PromptHandler
	serverInfo      mcp.ServerInfo
	logger          *slog.Logger
	config          *serverConfig
}

type serverConfig struct {
	requestTimeout  time.Duration
	shutdownTimeout time.Duration
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	logLevel        string
	logJSON         bool
	customLogger    *slog.Logger
}

type Option func(*serverConfig)

func WithLogger(logger *slog.Logger) Option {
	return func(cfg *serverConfig) {
		cfg.customLogger = logger
	}
}

func WithRequestTimeout(timeout time.Duration) Option {
	return func(cfg *serverConfig) {
		cfg.requestTimeout = timeout
	}
}

func WithShutdownTimeout(timeout time.Duration) Option {
	return func(cfg *serverConfig) {
		cfg.shutdownTimeout = timeout
	}
}

func WithReadTimeout(timeout time.Duration) Option {
	return func(cfg *serverConfig) {
		cfg.readTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) Option {
	return func(cfg *serverConfig) {
		cfg.writeTimeout = timeout
	}
}

func WithIdleTimeout(timeout time.Duration) Option {
	return func(cfg *serverConfig) {
		cfg.idleTimeout = timeout
	}
}

func WithLogLevel(level string) Option {
	return func(cfg *serverConfig) {
		cfg.logLevel = level
	}
}

func WithLogJSON(enabled bool) Option {
	return func(cfg *serverConfig) {
		cfg.logJSON = enabled
	}
}

// NewMCPServer creates a new MCP server using the options pattern.
//
// This constructor provides a more flexible way to configure the server
// using functional options. It requires the server name, version, and handlers,
// while all other settings can be configured via options.
//
// Example usage:
//
//	server, err := NewMCPServer(
//	    "My MCP Server", "1.0.0",
//	    toolHandler, resourceHandler, promptHandler,
//	    WithLogger(logger),
//	    WithRequestTimeout(30*time.Second),
//	    WithLogLevel("debug"),
//	)
func NewMCPServer(name, version string, toolHandler mcp.ToolHandler, resourceHandler mcp.ResourceHandler, promptHandler mcp.PromptHandler, opts ...Option) (*Server, error) {
	if toolHandler == nil {
		return nil, fmt.Errorf("toolHandler cannot be nil")
	}
	if resourceHandler == nil {
		return nil, fmt.Errorf("resourceHandler cannot be nil")
	}
	if promptHandler == nil {
		return nil, fmt.Errorf("promptHandler cannot be nil")
	}

	config := &serverConfig{
		requestTimeout:  30 * time.Second,
		shutdownTimeout: 5 * time.Second,
		readTimeout:     30 * time.Second,
		writeTimeout:    30 * time.Second,
		idleTimeout:     120 * time.Second,
		logLevel:        "info",
		logJSON:         false,
	}

	for _, opt := range opts {
		opt(config)
	}

	var logger *slog.Logger
	if config.customLogger != nil {
		logger = config.customLogger
	} else {
		logger = createDefaultLogger(config.logLevel, config.logJSON)
	}

	return &Server{
		toolHandler:     toolHandler,
		resourceHandler: resourceHandler,
		promptHandler:   promptHandler,
		logger:          logger,
		config:          config,
		serverInfo: mcp.ServerInfo{
			Name:    name,
			Version: version,
		},
	}, nil
}

func (s *Server) Initialize(ctx context.Context) (*mcp.InitializeResponse, error) {
	return &mcp.InitializeResponse{
		ProtocolVersion: mcp.ProtocolVersion,
		Capabilities: map[string]any{
			"tools":       map[string]bool{"listChanged": true},
			"resources":   map[string]bool{"listChanged": true, "templates": true},
			"prompts":     map[string]bool{"listChanged": true},
			"elicitation": map[string]any{},
		},
		ServerInfo: s.serverInfo,
	}, nil
}

func (s *Server) HandleRequest(ctx context.Context, req mcp.Request) error {
	s.logger.Debug("Handling request", "method", req.Method, "id", req.ID)

	switch req.Method {
	case "initialize":
		return s.handleInitialize(ctx, req.ID)
	case "tools/list":
		return s.handleToolsList(ctx, req.ID)
	case "tools/call":
		return s.handleToolsCall(ctx, req.ID, req)
	case "resources/list":
		return s.handleResourcesList(ctx, req.ID)
	case "resources/read":
		return s.handleResourcesRead(ctx, req.ID, req)
	case "resources/templates/list":
		return s.handleResourceTemplatesList(ctx, req.ID)
	case "prompts/list":
		return s.handlePromptsList(ctx, req.ID)
	case "prompts/get":
		return s.handlePromptsGet(ctx, req.ID, req)
	case "ping":
		return s.handlePing(ctx, req.ID)
	default:
		s.logger.Warn("Unknown method requested", "method", req.Method, "id", req.ID)
		return s.sendError(ctx, req.ID, mcp.ErrorCodeMethodNotFound, fmt.Sprintf("Method %s not found", req.Method), nil)
	}
}

func (s *Server) sendResponse(ctx context.Context, id any, result any) error {
	response := mcp.Response{
		JSONRPC: mcp.JSONRPCVersion,
		ID:      id,
		Result:  result,
	}
	return s.sendResponseDirect(ctx, response)
}

func (s *Server) sendError(ctx context.Context, id any, code int, message string, data any) error {
	sender := ctx.Value(mcp.ResponseSenderKey)
	if sender == nil {
		return fmt.Errorf("missing response sender in context")
	}

	rs, ok := sender.(mcp.ResponseSender)
	if !ok {
		return fmt.Errorf("invalid response sender type in context")
	}

	return rs.SendError(id, code, message, data)
}

// sendResponseDirect sends a JSON-RPC response directly.
func (s *Server) sendResponseDirect(ctx context.Context, response mcp.Response) error {
	sender := ctx.Value(mcp.ResponseSenderKey)
	if sender == nil {
		return fmt.Errorf("missing response sender in context")
	}

	rs, ok := sender.(mcp.ResponseSender)
	if !ok {
		return fmt.Errorf("invalid response sender type in context")
	}

	return rs.SendResponse(response)
}

// Request handlers
func (s *Server) handleInitialize(ctx context.Context, id any) error {
	result, err := s.Initialize(ctx)
	if err != nil {
		s.logger.Error("Failed to initialize server", "error", err, "id", id)
		return s.sendError(ctx, id, mcp.ErrorCodeInternalError, "Failed to initialize", err.Error())
	}
	s.logger.Info("Server initialized successfully", "id", id)
	return s.sendResponse(ctx, id, result)
}

func (s *Server) handleToolsList(ctx context.Context, id any) error {
	tools, err := s.toolHandler.ListTools(ctx)
	if err != nil {
		s.logger.Error("Failed to list tools", "error", err, "id", id)
		return s.sendError(ctx, id, mcp.ErrorCodeInternalError, "Failed to list tools", err.Error())
	}
	s.logger.Debug("Listed tools", "count", len(tools), "id", id)
	return s.sendResponse(ctx, id, map[string][]mcp.Tool{"tools": tools})
}

func (s *Server) handleToolsCall(ctx context.Context, id any, req mcp.Request) error {
	params, err := s.parseToolCallParams(req.Params)
	if err != nil {
		s.logger.Error("Invalid tool call parameters", "error", err, "id", id)
		return s.sendError(ctx, id, mcp.ErrorCodeInvalidParams, "Invalid tool call parameters", err.Error())
	}

	s.logger.Debug("Calling tool", "tool", params.Name, "id", id)
	response, err := s.toolHandler.CallTool(ctx, params)
	if err != nil {
		s.logger.Error("Tool call failed", "tool", params.Name, "error", err, "id", id)
		return s.sendError(ctx, id, mcp.ErrorCodeInvalidParams, fmt.Sprintf("Tool call failed: %s", err.Error()), nil)
	}
	s.logger.Debug("Tool call completed", "tool", params.Name, "id", id)
	return s.sendResponse(ctx, id, response)
}

func (s *Server) handleResourcesList(ctx context.Context, id any) error {
	resources, err := s.resourceHandler.ListResources(ctx)
	if err != nil {
		return s.sendError(ctx, id, mcp.ErrorCodeInternalError, "Failed to list resources", err.Error())
	}
	return s.sendResponse(ctx, id, map[string][]mcp.Resource{"resources": resources})
}

func (s *Server) handleResourcesRead(ctx context.Context, id any, req mcp.Request) error {
	params, err := s.parseResourceParams(req.Params)
	if err != nil {
		return s.sendError(ctx, id, mcp.ErrorCodeInvalidParams, "Invalid resource read parameters", err.Error())
	}

	response, err := s.resourceHandler.ReadResource(ctx, params)
	if err != nil {
		return s.sendError(ctx, id, mcp.ErrorCodeInvalidParams, fmt.Sprintf("Resource read failed: %s", err.Error()), nil)
	}
	return s.sendResponse(ctx, id, response)
}

func (s *Server) handleResourceTemplatesList(ctx context.Context, id any) error {
	templates, err := s.resourceHandler.ListResourceTemplates(ctx)
	if err != nil {
		s.logger.Error("Failed to list resource templates", "error", err, "id", id)
		return s.sendError(ctx, id, mcp.ErrorCodeInternalError, "Failed to list resource templates", err.Error())
	}
	s.logger.Debug("Listed resource templates", "count", len(templates), "id", id)
	return s.sendResponse(ctx, id, map[string][]mcp.ResourceTemplate{"resourceTemplates": templates})
}

func (s *Server) handlePromptsList(ctx context.Context, id any) error {
	prompts, err := s.promptHandler.ListPrompts(ctx)
	if err != nil {
		return s.sendError(ctx, id, mcp.ErrorCodeInternalError, "Failed to list prompts", err.Error())
	}
	return s.sendResponse(ctx, id, map[string][]mcp.Prompt{"prompts": prompts})
}

func (s *Server) handlePromptsGet(ctx context.Context, id any, req mcp.Request) error {
	params, err := s.parsePromptParams(req.Params)
	if err != nil {
		return s.sendError(ctx, id, mcp.ErrorCodeInvalidParams, "Invalid prompt parameters", err.Error())
	}

	response, err := s.promptHandler.GetPrompt(ctx, params)
	if err != nil {
		return s.sendError(ctx, id, mcp.ErrorCodeInvalidParams, fmt.Sprintf("Prompt call failed: %s", err.Error()), nil)
	}
	return s.sendResponse(ctx, id, response)
}

func (s *Server) handlePing(ctx context.Context, id any) error {
	return s.sendResponse(ctx, id, map[string]any{})
}

func (s *Server) parseToolCallParams(params any) (mcp.ToolCallParams, error) {
	if params == nil {
		return mcp.ToolCallParams{}, fmt.Errorf("params cannot be nil")
	}

	paramsMap, ok := params.(map[string]any)
	if !ok {
		return mcp.ToolCallParams{}, fmt.Errorf("params must be an object")
	}

	name, ok := paramsMap["name"].(string)
	if !ok {
		return mcp.ToolCallParams{}, fmt.Errorf("name parameter is required and must be a string")
	}

	args := make(map[string]any)
	if arguments, exists := paramsMap["arguments"]; exists {
		if argsMap, ok := arguments.(map[string]any); ok {
			args = argsMap
		}
	}

	return mcp.ToolCallParams{
		Name:      name,
		Arguments: args,
	}, nil
}

func (s *Server) parseResourceParams(params any) (mcp.ResourceParams, error) {
	if params == nil {
		return mcp.ResourceParams{}, fmt.Errorf("params cannot be nil")
	}

	paramsMap, ok := params.(map[string]any)
	if !ok {
		return mcp.ResourceParams{}, fmt.Errorf("params must be an object")
	}

	uri, ok := paramsMap["uri"].(string)
	if !ok {
		return mcp.ResourceParams{}, fmt.Errorf("uri parameter is required and must be a string")
	}

	return mcp.ResourceParams{URI: uri}, nil
}

func (s *Server) parsePromptParams(params any) (mcp.PromptParams, error) {
	if params == nil {
		return mcp.PromptParams{}, fmt.Errorf("params cannot be nil")
	}

	paramsMap, ok := params.(map[string]any)
	if !ok {
		return mcp.PromptParams{}, fmt.Errorf("params must be an object")
	}

	name, ok := paramsMap["name"].(string)
	if !ok {
		return mcp.PromptParams{}, fmt.Errorf("name parameter is required and must be a string")
	}

	args := make(map[string]any)
	if arguments, exists := paramsMap["arguments"]; exists {
		if argsMap, ok := arguments.(map[string]any); ok {
			args = argsMap
		}
	}

	return mcp.PromptParams{
		Name:      name,
		Arguments: args,
	}, nil
}

func createDefaultLogger(logLevel string, logJSON bool) *slog.Logger {
	var handler slog.Handler

	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	logOutput := os.Stderr

	log.SetOutput(os.Stderr)

	if logJSON {
		handler = slog.NewJSONHandler(logOutput, opts)
	} else {
		handler = slog.NewTextHandler(logOutput, opts)
	}

	return slog.New(handler)
}
