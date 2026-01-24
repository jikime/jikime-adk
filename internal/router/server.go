package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"jikime-adk/internal/router/provider"
)

// Server represents the LLM router proxy server.
type Server struct {
	config   *Config
	provider provider.Provider
	httpSrv  *http.Server
	logger   *log.Logger
}

// NewServer creates a new router server with the given configuration.
func NewServer(cfg *Config) (*Server, error) {
	provCfg, err := cfg.GetActiveProvider()
	if err != nil {
		return nil, err
	}

	pCfg := toProviderConfig(provCfg)
	prov, err := provider.NewProvider(cfg.Router.Provider, pCfg)
	if err != nil {
		return nil, err
	}

	logger := log.New(os.Stdout, "[router] ", log.LstdFlags)

	s := &Server{
		config:   cfg,
		provider: prov,
		logger:   logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages", s.handleMessages)
	mux.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf("%s:%d", cfg.Router.Host, cfg.Router.Port)
	s.httpSrv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return s, nil
}

// Start starts the proxy server and blocks until shutdown.
func (s *Server) Start() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s.logger.Printf("Starting on %s (provider: %s)",
			s.httpSrv.Addr, s.provider.Name())
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
	fmt.Fprintf(w, `{"status":"ok","provider":"%s"}`, s.provider.Name())
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
