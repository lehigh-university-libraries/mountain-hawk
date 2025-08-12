package github

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/v74/github"
)

// ParseWebhookEvent parses a GitHub webhook payload
func ParseWebhookEvent(r *http.Request, payload []byte) (interface{}, error) {
	eventType := github.WebHookType(r)
	if eventType == "" {
		return nil, fmt.Errorf("missing X-GitHub-Event header")
	}

	event, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	return event, nil
}

// IsSupportedEvent checks if the webhook event type is supported
func IsSupportedEvent(event interface{}) bool {
	switch event.(type) {
	case *github.PullRequestEvent:
		return true
	default:
		return false
	}
}

// GetPREventDetails extracts relevant information from a pull request event
func GetPREventDetails(event *github.PullRequestEvent) (action string, pr *github.PullRequest, repo *github.Repository) {
	return event.GetAction(), event.GetPullRequest(), event.GetRepo()
}

// ShouldProcessPREvent determines if a PR event should trigger a review
func ShouldProcessPREvent(action string) bool {
	switch action {
	case "opened", "synchronize", "reopened":
		return true
	default:
		return false
	}
}
