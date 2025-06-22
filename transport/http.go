package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cbrgm/go-mcp-server/mcp"
	"github.com/cbrgm/go-mcp-server/server"
)

const (
	contentTypeJSON = "application/json; charset=utf-8"
	contentTypeSSE  = "text/event-stream; charset=utf-8"
	contentTypeHTML = "text/html; charset=utf-8"

	headerMCPSessionID       = "Mcp-Session-Id"
	headerMCPProtocolVersion = "MCP-Protocol-Version"

	sessionIDPrefix = "session_"
)

type HTTPTransport struct {
	port            int
	server          *http.Server
	sessions        map[string]*SSESession
	mu              sync.RWMutex
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	shutdownTimeout time.Duration
	requestTimeout  time.Duration
}

type HTTPResponseSender struct {
	writer http.ResponseWriter
	sent   bool
	mu     sync.Mutex
}

func (h *HTTPResponseSender) SendResponse(response mcp.Response) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.sent {
		return fmt.Errorf("response already sent")
	}

	h.writer.Header().Set("Content-Type", contentTypeJSON)
	h.writer.WriteHeader(http.StatusOK)
	err := json.NewEncoder(h.writer).Encode(response)
	h.sent = true
	return err
}

func (h *HTTPResponseSender) SendError(id any, code int, message string, data any) error {
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
	return h.SendResponse(response)
}

type SSEResponseSender struct {
	session *SSESession
}

func (s *SSEResponseSender) SendResponse(response mcp.Response) error {
	return s.session.sendEvent("", response)
}

func (s *SSEResponseSender) SendError(id any, code int, message string, data any) error {
	return s.session.sendError(id, code, message, data)
}

type SSESession struct {
	ID      string
	writer  http.ResponseWriter
	flusher http.Flusher
	eventID int
	mu      sync.Mutex
	closed  bool
}

func NewHTTP(port int, readTimeout, writeTimeout, idleTimeout, shutdownTimeout, requestTimeout time.Duration) *HTTPTransport {
	return &HTTPTransport{
		port:            port,
		sessions:        make(map[string]*SSESession),
		readTimeout:     readTimeout,
		writeTimeout:    writeTimeout,
		idleTimeout:     idleTimeout,
		shutdownTimeout: shutdownTimeout,
		requestTimeout:  requestTimeout,
	}
}

func (t *HTTPTransport) Start(ctx context.Context, srv *server.Server) error {
	mux := http.NewServeMux()

	handler := t.corsMiddleware(t.securityMiddleware(mux))

	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			t.handlePost(ctx, srv, w, r)
		case http.MethodGet:
			t.handleGet(ctx, srv, w, r)
		case http.MethodOptions:
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		t.handleStatusPage(w, r)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("Content-Type", contentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	t.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", t.port),
		Handler:      handler,
		ReadTimeout:  t.readTimeout,
		WriteTimeout: t.writeTimeout,
		IdleTimeout:  t.idleTimeout,
	}

	log.Printf("Starting HTTP transport on port %d...", t.port)
	log.Printf("MCP endpoint: http://localhost:%d/mcp", t.port)

	go func() {
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("HTTP transport shutting down")
	return t.Stop()
}

func (t *HTTPTransport) Stop() error {
	t.mu.Lock()
	for _, session := range t.sessions {
		session.close()
	}
	t.sessions = make(map[string]*SSESession)
	t.mu.Unlock()

	if t.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), t.shutdownTimeout)
		defer cancel()
		return t.server.Shutdown(ctx)
	}
	return nil
}

func (t *HTTPTransport) handlePost(ctx context.Context, srv *server.Server, w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "application/json; charset=utf-8")

	// TODO: add back when uprading to the most recent MCP spec
	// protocolVersion := r.Header.Get("MCP-Protocol-Version")
	var req mcp.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		t.sendError(w, -1, mcp.ErrorCodeParseError, "Parse error", err.Error())
		return
	}

	// TODO: add back when uprading to the most recent MCP spec
	// if req.Method != "initialize" && protocolVersion != "" && protocolVersion != mcp.ProtocolVersion {
	// 	t.sendError(w, req.ID, mcp.ErrorCodeInvalidRequest,
	// 		fmt.Sprintf("Unsupported protocol version: %s", protocolVersion), nil)
	// 	return
	// }
	acceptHeader := r.Header.Get("Accept")
	wantsSSE := strings.Contains(acceptHeader, "text/event-stream")
	wantsJSON := strings.Contains(acceptHeader, "application/json")

	if !wantsJSON && !wantsSSE {
		t.sendError(w, req.ID, mcp.ErrorCodeInvalidRequest, "Accept header must include application/json and/or text/event-stream", nil)
		return
	}

	if req.JSONRPC != mcp.JSONRPCVersion {
		t.sendError(w, req.ID, mcp.ErrorCodeInvalidRequest, "Invalid JSON-RPC version", nil)
		return
	}

	// Handle notifications (no response expected)
	if req.ID == nil {
		log.Printf("Received notification: %s", req.Method)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// If client wants SSE and this is a request, start SSE stream
	if wantsSSE && req.ID != nil {
		t.handleSSERequest(ctx, srv, w, r, req)
		return
	}

	// Handle regular JSON response
	t.handleJSONRequest(ctx, srv, w, req)
}

func (t *HTTPTransport) handleGet(ctx context.Context, srv *server.Server, w http.ResponseWriter, r *http.Request) {
	_ = srv // Server not used for GET but kept for consistency
	// GET is used to open SSE streams or resume connections
	session := t.startSSEStream(w, r)
	if session == nil {
		return
	}

	// Keep the connection alive until context is cancelled
	<-ctx.Done()

	// Clean up session
	t.mu.Lock()
	delete(t.sessions, session.ID)
	t.mu.Unlock()
}

func (t *HTTPTransport) handleJSONRequest(ctx context.Context, srv *server.Server, w http.ResponseWriter, req mcp.Request) {
	reqCtx, cancel := context.WithTimeout(ctx, t.requestTimeout)
	defer cancel()

	httpSender := &HTTPResponseSender{writer: w}
	reqCtx = context.WithValue(reqCtx, mcp.ResponseSenderKey, httpSender)

	if err := srv.HandleRequest(reqCtx, req); err != nil {
		log.Printf("Error handling request: %v", err)
		if !httpSender.sent {
			t.sendError(w, req.ID, mcp.ErrorCodeInternalError, "Internal error", err.Error())
		}
		return
	}

	if !httpSender.sent {
		t.sendError(w, req.ID, mcp.ErrorCodeInternalError, "No response generated", nil)
	}
}

func (t *HTTPTransport) handleSSERequest(ctx context.Context, srv *server.Server, w http.ResponseWriter, r *http.Request, req mcp.Request) {
	session := t.startSSEStream(w, r)
	if session == nil {
		return
	}

	reqCtx, cancel := context.WithTimeout(ctx, t.requestTimeout)
	defer cancel()

	sseSender := &SSEResponseSender{session: session}
	reqCtx = context.WithValue(reqCtx, mcp.ResponseSenderKey, sseSender)
	reqCtx = context.WithValue(reqCtx, mcp.SessionIDKey, session.ID)

	if err := srv.HandleRequest(reqCtx, req); err != nil {
		log.Printf("Error handling SSE request: %v", err)
		session.sendError(req.ID, mcp.ErrorCodeInternalError, "Internal error", err.Error())
	}
}

func (t *HTTPTransport) startSSEStream(w http.ResponseWriter, r *http.Request) *SSESession {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return nil
	}

	w.Header().Set("Content-Type", contentTypeSSE)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	lastEventID := r.Header.Get("Last-Event-ID")
	eventID := 0
	if lastEventID != "" {
		if id, err := strconv.Atoi(lastEventID); err == nil {
			eventID = id + 1
		}
	}

	sessionID := r.Header.Get("Mcp-Session-Id")
	if sessionID == "" {
		sessionID = fmt.Sprintf("%s%d", sessionIDPrefix, time.Now().UnixNano())
	}
	session := &SSESession{
		ID:      sessionID,
		writer:  w,
		flusher: flusher,
		eventID: eventID,
	}

	t.mu.Lock()
	t.sessions[sessionID] = session
	t.mu.Unlock()

	w.Header().Set(headerMCPSessionID, sessionID)

	session.sendEvent("connected", map[string]string{
		"sessionId": sessionID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})

	return session
}

func (t *HTTPTransport) sendError(w http.ResponseWriter, id any, code int, message string, data any) {
	errorResp := mcp.Response{
		JSONRPC: mcp.JSONRPCVersion,
		ID:      id,
		Error: &mcp.ErrorResponse{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(errorResp)
}

func (s *SSESession) sendEvent(eventType string, data any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("session closed")
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	fmt.Fprintf(s.writer, "id: %d\n", s.eventID)
	if eventType != "" {
		fmt.Fprintf(s.writer, "event: %s\n", eventType)
	}

	dataStr := string(dataBytes)
	lines := strings.Split(dataStr, "\n")
	for _, line := range lines {
		fmt.Fprintf(s.writer, "data: %s\n", line)
	}
	fmt.Fprintf(s.writer, "\n")

	s.flusher.Flush()
	s.eventID++

	return nil
}

func (s *SSESession) sendError(id any, code int, message string, data any) error {
	errorResp := mcp.Response{
		JSONRPC: mcp.JSONRPCVersion,
		ID:      id,
		Error: &mcp.ErrorResponse{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	return s.sendEvent("", errorResp)
}

func (s *SSESession) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
}

func (t *HTTPTransport) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Last-Event-ID, Mcp-Session-Id, MCP-Protocol-Version")
		w.Header().Set("Access-Control-Allow-Credentials", "false")
		w.Header().Set("Access-Control-Max-Age", "86400")

		w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id, MCP-Protocol-Version")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (t *HTTPTransport) handleStatusPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeHTML)
	w.WriteHeader(http.StatusOK)

	t.mu.RLock()
	activeSessions := len(t.sessions)
	t.mu.RUnlock()

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MCP Server</title>
    <style>
        * { box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 0;
            background: #f8f9fa;
            color: #2c3e50;
            line-height: 1.6;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            padding: 3rem 2rem;
        }
        .header {
            text-align: center;
            margin-bottom: 3rem;
        }
        .header h1 {
            margin: 0 0 0.5rem 0;
            font-size: 2rem;
            font-weight: 300;
            color: #2c3e50;
        }
        .header p {
            margin: 0;
            color: #6c757d;
            font-size: 1rem;
        }
        .status {
            background: #d1ecf1;
            color: #0c5460;
            padding: 1rem 1.5rem;
            border-radius: 6px;
            margin-bottom: 2rem;
            text-align: center;
            font-weight: 500;
        }
        .info {
            background: white;
            border-radius: 6px;
            padding: 1.5rem;
            margin-bottom: 2rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .info-row {
            display: flex;
            justify-content: space-between;
            padding: 0.5rem 0;
            border-bottom: 1px solid #e9ecef;
        }
        .info-row:last-child { border-bottom: none; }
        .label { color: #6c757d; }
        .value {
            font-family: 'Monaco', 'Consolas', monospace;
            color: #2c3e50;
            font-size: 0.9rem;
        }
        .endpoints {
            background: white;
            border-radius: 6px;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .endpoints h3 {
            margin: 0 0 1rem 0;
            font-size: 1.1rem;
            color: #2c3e50;
        }
        .endpoint {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0.75rem 0;
            border-bottom: 1px solid #e9ecef;
            font-family: 'Monaco', 'Consolas', monospace;
            font-size: 0.9rem;
        }
        .endpoint:last-child { border-bottom: none; }
        .method {
            background: #007bff;
            color: white;
            padding: 0.2rem 0.5rem;
            border-radius: 3px;
            font-size: 0.75rem;
            font-weight: bold;
            margin-right: 0.5rem;
        }
        .footer {
            text-align: center;
            margin-top: 2rem;
            padding-top: 2rem;
            border-top: 1px solid #e9ecef;
            color: #6c757d;
            font-size: 0.9rem;
        }
        .footer a {
            color: #007bff;
            text-decoration: none;
        }
        .footer a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>MCP Server</h1>
            <p>Model Context Protocol</p>
        </div>

        <div class="status">
            Running on port %d
        </div>

        <div class="info">
            <div class="info-row">
                <span class="label">Protocol</span>
                <span class="value">%s</span>
            </div>
            <div class="info-row">
                <span class="label">Transport</span>
                <span class="value">HTTP + SSE</span>
            </div>
            <div class="info-row">
                <span class="label">Active Sessions</span>
                <span class="value">%d</span>
            </div>
        </div>

        <div class="endpoints">
            <h3>Endpoints</h3>
            <div class="endpoint">
                <div><span class="method">POST</span>/mcp</div>
                <span>JSON-RPC 2.0</span>
            </div>
            <div class="endpoint">
                <div><span class="method">GET</span>/mcp</div>
                <span>Server-Sent Events</span>
            </div>
            <div class="endpoint">
                <div><span class="method">GET</span>/health</div>
                <span>Health Check</span>
            </div>
        </div>

        <div class="footer">
            <a href="https://github.com/cbrgm/go-mcp-server">github.com/cbrgm/go-mcp-server</a>
        </div>
    </div>
</body>
</html>`

	fmt.Fprintf(w, html,
		t.port,              // Port
		mcp.ProtocolVersion, // MCP protocol version
		activeSessions,      // Active sessions
	)
}

func (t *HTTPTransport) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		next.ServeHTTP(w, r)
	})
}
