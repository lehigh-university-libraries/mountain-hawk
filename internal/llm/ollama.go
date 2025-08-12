package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lehigh-university-libraries/mountain-hawk/pkg/types"
)

// OllamaClient implements the Client interface for Ollama
type OllamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for code analysis
		},
	}
}

// ReviewCode sends code for review to Ollama
func (c *OllamaClient) ReviewCode(ctx context.Context, prompt string) (*types.ReviewResponse, error) {
	// Create the request
	reqBody := OllamaRequest{
		Model:  c.model,
		Prompt: c.buildReviewPrompt(prompt),
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to Ollama
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	// Parse Ollama response
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if ollamaResp.Error != "" {
		return nil, fmt.Errorf("ollama error: %s", ollamaResp.Error)
	}

	// Parse the review response
	return c.parseReviewResponse(ollamaResp.Response)
}

// GetModel returns the model being used
func (c *OllamaClient) GetModel() string {
	return c.model
}

// Health checks if Ollama is available
func (c *OllamaClient) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ollama health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama health check returned status %d", resp.StatusCode)
	}

	return nil
}

// buildReviewPrompt creates the structured prompt for code review
func (c *OllamaClient) buildReviewPrompt(context string) string {
	return fmt.Sprintf(`You are a code reviewer. Review the following pull request and return ONLY valid, parsable JSON matching exactly the schema below. Do not include any additional text, comments, or code fences.

Schema:
{
  "decision": "approve|request_changes|comment",
  "decision_rationale": "Brief explanation of approval/rejection",
  "general_comments": [
    {
      "body": "Overall feedback about the PR",
      "severity": "info|warning|error"
    }
  ],
  "file_comments": [
    {
      "path": "exact/file/path.ext",
      "line": <integer, 1-based line number from NEW file version>,
      "body": "Specific feedback for this line",
      "severity": "info|warning|error",
      "type": "bug|style|performance|security|maintainability"
    }
  ],
  "summary": "Brief summary of the review"
}

Guidelines:
- Use exact file paths from the PR.
- Line numbers must match the new file content (after changes).
- Only include file comments for lines that need feedback.
- Use "error" severity for bugs or security issues, "warning" for best practices, "info" for suggestions.
- Escape all quotes and special characters inside JSON strings.
- Focus on: security vulnerabilities, bugs, performance issues, maintainability
- Be constructive and specific in feedback

Context:
%s

Respond ONLY with raw JSON matching the schema above. Do not wrap it in backticks or other formatting.`, context)
}

// parseReviewResponse parses the LLM response into structured review data
func (c *OllamaClient) parseReviewResponse(response string) (*types.ReviewResponse, error) {
	var reviewResp types.ReviewResponse

	// Try to parse directly first
	if err := json.Unmarshal([]byte(response), &reviewResp); err != nil {
		// Fallback: try to extract JSON from response if model added extra text
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}")

		if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
			jsonStr := response[jsonStart : jsonEnd+1]
			if err := json.Unmarshal([]byte(jsonStr), &reviewResp); err != nil {
				return nil, fmt.Errorf("failed to parse model response as JSON: %w\nResponse: %s", err, response)
			}
		} else {
			return nil, fmt.Errorf("model response is not valid JSON: %s", response)
		}
	}

	// Validate the response
	if err := c.validateReviewResponse(&reviewResp); err != nil {
		return nil, fmt.Errorf("invalid review response: %w", err)
	}

	return &reviewResp, nil
}

// validateReviewResponse ensures the response has valid values
func (c *OllamaClient) validateReviewResponse(review *types.ReviewResponse) error {
	// Validate decision
	switch review.Decision {
	case types.DecisionApprove, types.DecisionRequestChanges, types.DecisionComment:
		// Valid
	default:
		return fmt.Errorf("invalid decision: %s", review.Decision)
	}

	// Validate general comments
	for i, comment := range review.GeneralComments {
		if err := c.validateSeverity(comment.Severity); err != nil {
			return fmt.Errorf("general comment %d: %w", i, err)
		}
	}

	// Validate file comments
	for i, comment := range review.FileComments {
		if comment.Path == "" {
			return fmt.Errorf("file comment %d: missing path", i)
		}
		if comment.Line <= 0 {
			return fmt.Errorf("file comment %d: invalid line number %d", i, comment.Line)
		}
		if err := c.validateSeverity(comment.Severity); err != nil {
			return fmt.Errorf("file comment %d: %w", i, err)
		}
		if err := c.validateCommentType(comment.Type); err != nil {
			return fmt.Errorf("file comment %d: %w", i, err)
		}
	}

	return nil
}

// validateSeverity checks if severity is valid
func (c *OllamaClient) validateSeverity(severity types.Severity) error {
	switch severity {
	case types.SeverityInfo, types.SeverityWarning, types.SeverityError:
		return nil
	default:
		return fmt.Errorf("invalid severity: %s", severity)
	}
}

// validateCommentType checks if comment type is valid
func (c *OllamaClient) validateCommentType(commentType types.CommentType) error {
	switch commentType {
	case types.TypeBug, types.TypeStyle, types.TypePerformance, types.TypeSecurity, types.TypeMaintainability:
		return nil
	default:
		return fmt.Errorf("invalid comment type: %s", commentType)
	}
}
