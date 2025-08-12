package server

import (
	"net/http"

	"github.com/lehigh-university-libraries/mountain-hawk/internal/config"
	"github.com/lehigh-university-libraries/mountain-hawk/internal/github"
	"github.com/lehigh-university-libraries/mountain-hawk/internal/llm"
	"github.com/lehigh-university-libraries/mountain-hawk/internal/reviewer"
)

// Server holds the HTTP server and its dependencies
type Server struct {
	config        *config.Config
	reviewService *reviewer.Service
	mux           *http.ServeMux
}

// New creates a new server with all dependencies initialized
func New(cfg *config.Config) *Server {
	// Initialize dependencies
	githubClient := github.NewClient(cfg.GitHubToken)
	llmClient := llm.NewOllamaClient(cfg.OllamaURL, cfg.OllamaModel)
	reviewService := reviewer.NewService(githubClient, llmClient)

	// Create server
	s := &Server{
		config:        cfg,
		reviewService: reviewService,
		mux:           http.NewServeMux(),
	}

	// Setup routes
	s.setupRoutes()

	return s
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/webhook", s.handleWebhook)
	s.mux.HandleFunc("/health", s.handleHealth)
}

// ListenAndServe starts the HTTP server
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(":"+s.config.Port, s.mux)
}
