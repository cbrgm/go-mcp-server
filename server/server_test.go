package server

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/cbrgm/go-mcp-server/cmd/go-mcp-server/handlers"
	"github.com/cbrgm/go-mcp-server/mcp"
)

func TestNewMCPServerWithOptions(t *testing.T) {
	// Create a handler for testing
	handler := &handlers.TeaHandler{}

	// Create a custom logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Test creating server with options
	server, err := NewMCPServer(
		"Test Server",
		"1.0.0",
		handler, handler, handler,
		WithLogger(logger),
		WithRequestTimeout(45*time.Second),
		WithLogLevel("debug"),
		WithLogJSON(true),
		WithReadTimeout(25*time.Second),
		WithWriteTimeout(25*time.Second),
		WithIdleTimeout(90*time.Second),
		WithShutdownTimeout(15*time.Second),
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	// Verify server info
	if server.serverInfo.Name != "Test Server" {
		t.Errorf("Expected server name 'Test Server', got '%s'", server.serverInfo.Name)
	}

	if server.serverInfo.Version != "1.0.0" {
		t.Errorf("Expected server version '1.0.0', got '%s'", server.serverInfo.Version)
	}

	// Verify config was applied
	if server.config.requestTimeout != 45*time.Second {
		t.Errorf("Expected request timeout 45s, got %v", server.config.requestTimeout)
	}

	if server.config.logLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", server.config.logLevel)
	}

	if !server.config.logJSON {
		t.Error("Expected logJSON to be true")
	}

	if server.config.readTimeout != 25*time.Second {
		t.Errorf("Expected read timeout 25s, got %v", server.config.readTimeout)
	}

	if server.config.writeTimeout != 25*time.Second {
		t.Errorf("Expected write timeout 25s, got %v", server.config.writeTimeout)
	}

	if server.config.idleTimeout != 90*time.Second {
		t.Errorf("Expected idle timeout 90s, got %v", server.config.idleTimeout)
	}

	if server.config.shutdownTimeout != 15*time.Second {
		t.Errorf("Expected shutdown timeout 15s, got %v", server.config.shutdownTimeout)
	}

	// Verify logger was set
	if server.logger != logger {
		t.Error("Expected custom logger to be set")
	}
}

func TestNewMCPServerDefaults(t *testing.T) {
	// Create a handler for testing
	handler := &handlers.TeaHandler{}

	// Test creating server with no options (should use defaults)
	server, err := NewMCPServer(
		"Default Server",
		"2.0.0",
		handler, handler, handler,
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	// Verify defaults
	if server.config.requestTimeout != 30*time.Second {
		t.Errorf("Expected default request timeout 30s, got %v", server.config.requestTimeout)
	}

	if server.config.shutdownTimeout != 5*time.Second {
		t.Errorf("Expected default shutdown timeout 5s, got %v", server.config.shutdownTimeout)
	}

	if server.config.logLevel != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", server.config.logLevel)
	}

	if server.config.logJSON {
		t.Error("Expected default logJSON to be false")
	}

	// Verify a default logger was created
	if server.logger == nil {
		t.Error("Expected default logger to be created")
	}
}

func TestNewMCPServerValidation(t *testing.T) {
	handler := &handlers.TeaHandler{}

	tests := []struct {
		name            string
		toolHandler     mcp.ToolHandler
		resourceHandler mcp.ResourceHandler
		promptHandler   mcp.PromptHandler
		expectError     bool
	}{
		{"nil tool handler", nil, handler, handler, true},
		{"nil resource handler", handler, nil, handler, true},
		{"nil prompt handler", handler, handler, nil, true},
		{"all handlers valid", handler, handler, handler, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMCPServer("Test", "1.0.0", tt.toolHandler, tt.resourceHandler, tt.promptHandler)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
