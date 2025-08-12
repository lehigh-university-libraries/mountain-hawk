package github

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/v74/github"
	"github.com/lehigh-university-libraries/mountain-hawk/pkg/types"
)

// ReviewPoster handles posting review results to GitHub
type ReviewPoster struct {
	client *Client
}

// NewReviewPoster creates a new review poster
func NewReviewPoster(client *Client) *ReviewPoster {
	return &ReviewPoster{client: client}
}

// PostReview posts a structured review to GitHub
func (rp *ReviewPoster) PostReview(ctx context.Context, owner, repo string, prNumber int, review *types.ReviewResponse, files []*github.CommitFile) error {
	// Create file map for position calculation
	fileMap := make(map[string]*github.CommitFile)
	for _, file := range files {
		fileMap[file.GetFilename()] = file
	}

	// Build review comments
	var reviewComments []*github.DraftReviewComment
	var generalBody strings.Builder

	// Add general comments to review body
	for _, comment := range review.GeneralComments {
		if generalBody.Len() > 0 {
			generalBody.WriteString("\n\n")
		}
		generalBody.WriteString(rp.formatGeneralComment(comment))
	}

	// Add summary if provided
	if review.Summary != "" {
		if generalBody.Len() > 0 {
			generalBody.WriteString("\n\n---\n\n")
		}
		generalBody.WriteString(fmt.Sprintf("**Summary:** %s", review.Summary))
	}

	// Process file comments
	for _, comment := range review.FileComments {
		file, exists := fileMap[comment.Path]
		if !exists {
			log.Printf("File %s not found in PR files", comment.Path)
			continue
		}

		// Calculate diff position
		position := CalculateDiffPosition(file, comment.Line)
		if position == -1 {
			log.Printf("Could not calculate position for %s:%d", comment.Path, comment.Line)
			continue
		}

		body := rp.formatFileComment(comment)
		reviewComments = append(reviewComments, &github.DraftReviewComment{
			Path:     &comment.Path,
			Position: &position,
			Body:     &body,
		})
	}

	// Create the review request
	reviewRequest := &github.PullRequestReviewRequest{
		Event:    github.String(rp.mapDecisionToEvent(review.Decision)),
		Comments: reviewComments,
	}

	if generalBody.Len() > 0 {
		body := generalBody.String()
		reviewRequest.Body = &body
	}

	// Post the review
	err := rp.client.CreateReview(ctx, owner, repo, prNumber, reviewRequest)
	if err != nil {
		// Fallback: post as general comment
		if generalBody.Len() > 0 {
			log.Printf("Failed to create review, posting as comment: %v", err)
			return rp.client.CreateIssueComment(ctx, owner, repo, prNumber, generalBody.String())
		}
		return err
	}

	return nil
}

// formatGeneralComment formats a general comment with appropriate emoji
func (rp *ReviewPoster) formatGeneralComment(comment types.GeneralComment) string {
	severity := rp.getSeverityEmoji(comment.Severity)
	return fmt.Sprintf("%s%s", severity, comment.Body)
}

// formatFileComment formats a file comment with severity and type indicators
func (rp *ReviewPoster) formatFileComment(comment types.FileComment) string {
	severity := rp.getSeverityEmoji(comment.Severity)
	typeEmoji := rp.getTypeEmoji(comment.Type)

	return fmt.Sprintf("%s%s**%s**: %s",
		severity,
		typeEmoji,
		strings.Title(string(comment.Type)),
		comment.Body,
	)
}

// getSeverityEmoji returns emoji for severity level
func (rp *ReviewPoster) getSeverityEmoji(severity types.Severity) string {
	switch severity {
	case types.SeverityError:
		return "🚨 "
	case types.SeverityWarning:
		return "⚠️ "
	case types.SeverityInfo:
		return "💡 "
	default:
		return ""
	}
}

// getTypeEmoji returns emoji for comment type
func (rp *ReviewPoster) getTypeEmoji(commentType types.CommentType) string {
	switch commentType {
	case types.TypeBug:
		return "🐛 "
	case types.TypeSecurity:
		return "🔒 "
	case types.TypePerformance:
		return "⚡ "
	case types.TypeStyle:
		return "🎨 "
	case types.TypeMaintainability:
		return "🔧 "
	default:
		return ""
	}
}

// mapDecisionToEvent maps review decision to GitHub review event
func (rp *ReviewPoster) mapDecisionToEvent(decision types.ReviewDecision) string {
	switch decision {
	case types.DecisionApprove:
		return "APPROVE"
	case types.DecisionRequestChanges:
		return "REQUEST_CHANGES"
	default:
		return "COMMENT"
	}
}
