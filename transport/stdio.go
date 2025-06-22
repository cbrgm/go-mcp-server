package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cbrgm/go-mcp-server/mcp"
	"github.com/cbrgm/go-mcp-server/server"
)

const (
	DefaultStdioTimeout = 30 * time.Second
)

type Stdio struct{}

func NewStdio() *Stdio {
	return &Stdio{}
}

func (t *Stdio) Start(ctx context.Context, srv *server.Server) error {
	log.Println("Starting stdio transport...")

	scanner := bufio.NewScanner(os.Stdin)

	lineChan := make(chan string)
	errChan := make(chan error)

	go func() {
		defer close(lineChan)
		defer close(errChan)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case lineChan <- scanner.Text():
			}
		}

		if err := scanner.Err(); err != nil {
			select {
			case <-ctx.Done():
				return
			case errChan <- err:
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stdio transport shutting down")
			return nil
		case err := <-errChan:
			if err != nil {
				log.Printf("Error reading input: %v", err)
			}
			return err
		case line, ok := <-lineChan:
			if !ok {
				log.Println("Input closed, exiting")
				return nil
			}

			if line == "" {
				continue
			}

			if err := t.handleMessage(ctx, srv, line); err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}
	}
}

func (t *Stdio) Stop() error {
	return nil
}

func (t *Stdio) handleMessage(ctx context.Context, srv *server.Server, line string) error {
	var req mcp.Request
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		return t.sendParseError(line, err)
	}

	if req.JSONRPC != mcp.JSONRPCVersion {
		log.Printf("Invalid JSON-RPC version: %s", req.JSONRPC)
		return nil
	}

	if req.ID == nil {
		log.Printf("Received notification: %s", req.Method)
		return nil
	}

	reqCtx := context.WithValue(ctx, mcp.ResponseSenderKey, &StdoutSender{})
	reqCtx, cancel := context.WithTimeout(reqCtx, DefaultStdioTimeout)
	defer cancel()

	return srv.HandleRequest(reqCtx, req)
}

func (t *Stdio) sendParseError(line string, err error) error {
	errorID := any(-1)
	var partialReq map[string]any
	if unmarshalErr := json.Unmarshal([]byte(line), &partialReq); unmarshalErr == nil {
		if id, exists := partialReq["id"]; exists && id != nil {
			errorID = id
		}
	}

	errorResp := mcp.Response{
		JSONRPC: mcp.JSONRPCVersion,
		ID:      errorID,
		Error: &mcp.ErrorResponse{
			Code:    mcp.ErrorCodeParseError,
			Message: "Parse error",
			Data:    err.Error(),
		},
	}

	respBytes, marshErr := json.Marshal(errorResp)
	if marshErr != nil {
		return marshErr
	}

	fmt.Println(string(respBytes))
	return nil
}

type StdoutSender struct{}

func (s *StdoutSender) SendResponse(response mcp.Response) error {
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	fmt.Println(string(jsonBytes))
	return nil
}

func (s *StdoutSender) SendError(id any, code int, message string, data any) error {
	errorResp := &mcp.ErrorResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
	response := mcp.Response{
		JSONRPC: mcp.JSONRPCVersion,
		ID:      id,
		Error:   errorResp,
	}
	return s.SendResponse(response)
}
