package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/cbrgm/go-mcp-server/cmd/go-mcp-server/handlers"
	"github.com/cbrgm/go-mcp-server/server"
	"github.com/cbrgm/go-mcp-server/transport"
)

const (
	transportStdio = "stdio"
	transportHTTP  = "http"

	defaultServerName    = "MCP Server"
	defaultServerVersion = "1.0.0"
	defaultHTTPPort      = 8080

	minPort = 1
	maxPort = 65535
)

type Config struct {
	TransportType   string        `arg:"--transport,env:MCP_TRANSPORT" default:"stdio" help:"Transport type (stdio|http)"`
	HTTPPort        int           `arg:"--port,env:MCP_PORT" default:"8080" help:"HTTP port"`
	ServerName      string        `arg:"--name,env:MCP_SERVER_NAME" default:"MCP Server" help:"Server name"`
	ServerVersion   string        `arg:"--version,env:MCP_SERVER_VERSION" default:"1.0.0" help:"Server version"`
	RequestTimeout  time.Duration `arg:"--request-timeout,env:MCP_REQUEST_TIMEOUT" default:"30s" help:"Request timeout"`
	ShutdownTimeout time.Duration `arg:"--shutdown-timeout,env:MCP_SHUTDOWN_TIMEOUT" default:"5s" help:"Shutdown timeout"`
	ReadTimeout     time.Duration `arg:"--read-timeout,env:MCP_READ_TIMEOUT" default:"30s" help:"HTTP read timeout"`
	WriteTimeout    time.Duration `arg:"--write-timeout,env:MCP_WRITE_TIMEOUT" default:"30s" help:"HTTP write timeout"`
	IdleTimeout     time.Duration `arg:"--idle-timeout,env:MCP_IDLE_TIMEOUT" default:"120s" help:"HTTP idle timeout"`
	LogLevel        string        `arg:"--log-level,env:MCP_LOG_LEVEL" default:"info" help:"Log level (debug|info|warn|error)"`
	LogJSON         bool          `arg:"--log-json,env:MCP_LOG_JSON" help:"Output logs in JSON format"`
}

func (Config) Description() string {
	return `MCP Server - A Model Context Protocol server implementation

This application provides a sample MCP server implementation that demonstrates
tools, resources, and prompts through the Model Context Protocol (MCP). 
It supports both stdio and HTTP transports for integration with various MCP clients.

Configuration can be provided via command line arguments or environment variables.
Environment variables use the prefix "MCP_" followed by the uppercase field name.

Examples:
  # Run with stdio transport (default)
  go-mcp-server

  # Run with HTTP transport on port 3000
  go-mcp-server --transport http --port 3000

  # Set server name via environment variable
  MCP_SERVER_NAME="My MCP Server" go-mcp-server`
}

func (Config) Version() string {
	return "go-mcp-server 1.0.0"
}

func (c *Config) Validate() error {
	switch c.TransportType {
	case transportStdio, transportHTTP:
	default:
		return fmt.Errorf("invalid transport type: %s (must be '%s' or '%s')", c.TransportType, transportStdio, transportHTTP)
	}

	if c.HTTPPort < minPort || c.HTTPPort > maxPort {
		return fmt.Errorf("invalid port: %d (must be %d-%d)", c.HTTPPort, minPort, maxPort)
	}

	if c.RequestTimeout <= 0 {
		return fmt.Errorf("invalid request timeout: %v (must be positive)", c.RequestTimeout)
	}

	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("invalid shutdown timeout: %v (must be positive)", c.ShutdownTimeout)
	}

	if c.ReadTimeout <= 0 {
		return fmt.Errorf("invalid read timeout: %v (must be positive)", c.ReadTimeout)
	}

	if c.WriteTimeout <= 0 {
		return fmt.Errorf("invalid write timeout: %v (must be positive)", c.WriteTimeout)
	}

	if c.IdleTimeout <= 0 {
		return fmt.Errorf("invalid idle timeout: %v (must be positive)", c.IdleTimeout)
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid log level: %s (must be 'debug', 'info', 'warn', or 'error')", c.LogLevel)
	}

	return nil
}

func parseArgs() (*Config, error) {
	var cfg Config

	parser, err := arg.NewParser(arg.Config{
		Program: "go-mcp-server",
	}, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create argument parser: %w", err)
	}

	err = parser.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

func main() {
	cfg, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func run(cfg *Config) error {
	teaHandler := &handlers.TeaHandler{}

	mcpServer, err := server.NewMCPServer(
		cfg.ServerName,
		cfg.ServerVersion,
		teaHandler, teaHandler, teaHandler,
		server.WithRequestTimeout(cfg.RequestTimeout),
		server.WithShutdownTimeout(cfg.ShutdownTimeout),
		server.WithReadTimeout(cfg.ReadTimeout),
		server.WithWriteTimeout(cfg.WriteTimeout),
		server.WithIdleTimeout(cfg.IdleTimeout),
		server.WithLogLevel(cfg.LogLevel),
		server.WithLogJSON(cfg.LogJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	transport, err := createTransport(cfg)
	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	if err := transport.Start(ctx, mcpServer); err != nil {
		return fmt.Errorf("transport start failed: %w", err)
	}

	return nil
}

func createTransport(cfg *Config) (transport.Transport, error) {
	switch strings.ToLower(cfg.TransportType) {
	case transportStdio:
		return transport.NewStdio(), nil
	case transportHTTP:
		return transport.NewHTTP(cfg.HTTPPort, cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout, cfg.ShutdownTimeout, cfg.RequestTimeout), nil
	default:
		return nil, fmt.Errorf("invalid transport type: %s (must be '%s' or '%s')", cfg.TransportType, transportStdio, transportHTTP)
	}
}
