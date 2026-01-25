package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"jikime-adk/internal/router/provider"
)

// Server represents the LLM router proxy server.
type Server struct {
	config  *Config
	httpSrv *http.Server
	logger  *log.Logger
}

// NewServer creates a new router server with the given configuration.
func NewServer(cfg *Config) (*Server, error) {
	if len(cfg.Providers) == 0 {
		return nil, fmt.Errorf("no providers configured")
	}

	logger := log.New(os.Stdout, "[router] ", log.LstdFlags)

	s := &Server{
		config: cfg,
		logger: logger,
	}

	mux := http.NewServeMux()
	// Route pattern: /{provider}/v1/messages
	mux.HandleFunc("/", s.routeHandler)

	addr := fmt.Sprintf("%s:%d", cfg.Router.Host, cfg.Router.Port)
	s.httpSrv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return s, nil
}

// routeHandler routes requests based on URL path.
// Expected paths:
//   - /{provider}/v1/messages - API messages endpoint
//   - /health - Health check
func (s *Server) routeHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Health check
	if path == "/health" {
		s.handleHealth(w, r)
		return
	}

	// Ignore Claude Code internal endpoints (event logging, telemetry, etc.)
	if strings.HasPrefix(path, "/api/") {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse provider from path: /{provider}/v1/messages
	parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
	if len(parts) < 2 {
		s.logger.Printf("[ERR] Invalid path: %s", path)
		s.writeError(w, http.StatusNotFound, "invalid_request_error", "Invalid path")
		return
	}

	providerName := parts[0]
	subPath := "/" + parts[1]

	// Validate provider exists
	if _, ok := s.config.Providers[providerName]; !ok {
		s.logger.Printf("[ERR] Unknown provider: %s", providerName)
		s.writeError(w, http.StatusNotFound, "invalid_request_error",
			fmt.Sprintf("Unknown provider: %s", providerName))
		return
	}

	// Route to appropriate handler
	if subPath == "/v1/messages" {
		s.handleMessages(w, r, providerName)
		return
	}

	s.logger.Printf("[ERR] Endpoint not found: %s", path)
	s.writeError(w, http.StatusNotFound, "invalid_request_error", "Endpoint not found")
}

// Start starts the proxy server and blocks until shutdown.
func (s *Server) Start() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		providers := s.config.GetProviderNames()
		s.logger.Printf("Starting on %s (providers: %s)",
			s.httpSrv.Addr, strings.Join(providers, ", "))
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	s.logger.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5_000_000_000) // 5s
	defer cancel()

	return s.httpSrv.Shutdown(ctx)
}

// Addr returns the server's listen address.
func (s *Server) Addr() string {
	return s.httpSrv.Addr
}

// handleHealth responds to health check requests.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	providers := s.config.GetProviderNames()
	resp := map[string]any{
		"status":    "ok",
		"providers": providers,
	}
	json.NewEncoder(w).Encode(resp)
}

// toProviderConfig converts router config to provider config.
func toProviderConfig(cfg *ProviderConfig) *provider.ProviderConfig {
	return &provider.ProviderConfig{
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		BaseURL: cfg.BaseURL,
		Region:  cfg.Region,
	}
}
