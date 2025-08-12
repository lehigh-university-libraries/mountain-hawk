package llm

import (
	"context"

	"github.com/lehigh-university-libraries/mountain-hawk/pkg/types"
)

// Client defines the interface for LLM providers
type Client interface {
	// ReviewCode sends code for review and returns structured feedback
	ReviewCode(ctx context.Context, prompt string) (*types.ReviewResponse, error)

	// GetModel returns the model being used
	GetModel() string

	// Health checks if the LLM service is available
	Health(ctx context.Context) error
}
