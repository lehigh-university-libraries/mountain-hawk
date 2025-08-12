package server

import (
	"io"
	"log"
	"net/http"

	"github.com/google/go-github/v74/github"
)

// handleWebhook processes GitHub webhook events
func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read webhook body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Parse webhook event
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("Failed to parse webhook: %v", err)
		http.Error(w, "Failed to parse webhook", http.StatusBadRequest)
		return
	}

	// Handle different event types
	switch e := event.(type) {
	case *github.PullRequestEvent:
		s.handlePullRequestEvent(e)
	default:
		log.Printf("Unhandled event type: %T", event)
	}

	w.WriteHeader(http.StatusOK)
}

// handlePullRequestEvent processes pull request events
func (s *Server) handlePullRequestEvent(event *github.PullRequestEvent) {
	action := event.GetAction()

	// Only process opened and synchronized (updated) PRs
	if action != "opened" && action != "synchronize" {
		log.Printf("Ignoring PR action: %s", action)
		return
	}

	pr := event.GetPullRequest()
	repo := event.GetRepo()

	log.Printf("Processing PR #%d: %s", pr.GetNumber(), pr.GetTitle())

	// Process review asynchronously to avoid webhook timeouts
	go func() {
		if err := s.reviewService.ReviewPR(pr, repo); err != nil {
			log.Printf("Failed to review PR #%d: %v", pr.GetNumber(), err)
		}
	}()
}

// handleHealth provides a health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
