package reviewer

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v74/github"
	gh "github.com/lehigh-university-libraries/mountain-hawk/internal/github"
	"github.com/lehigh-university-libraries/mountain-hawk/internal/llm"
	"github.com/lehigh-university-libraries/mountain-hawk/pkg/types"
)

// Service orchestrates the PR review process
type Service struct {
	githubClient   *gh.Client
	llmClient      llm.Client
	contextBuilder *ContextBuilder
	reviewPoster   *gh.ReviewPoster
}

// NewService creates a new reviewer service
func NewService(githubClient *gh.Client, llmClient llm.Client) *Service {
	return &Service{
		githubClient:   githubClient,
		llmClient:      llmClient,
		contextBuilder: NewContextBuilder(githubClient),
		reviewPoster:   gh.NewReviewPoster(githubClient),
	}
}

// ReviewPR performs a complete review of a pull request
func (s *Service) ReviewPR(pr *github.PullRequest, repo *github.Repository) error {
	ctx := context.Background()

	owner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()
	prNumber := pr.GetNumber()

	log.Printf("Starting review for PR #%d in %s/%s", prNumber, owner, repoName)

	// Get PR files
	files, err := s.githubClient.GetPRFiles(ctx, owner, repoName, prNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR files: %w", err)
	}

	if len(files) == 0 {
		log.Printf("No files to review in PR #%d", prNumber)
		return nil
	}

	// Build context for LLM
	context, err := s.contextBuilder.BuildContext(ctx, repo, pr, files)
	if err != nil {
		return fmt.Errorf("failed to build context: %w", err)
	}

	// Get review from LLM
	review, err := s.llmClient.ReviewCode(ctx, context)
	if err != nil {
		return fmt.Errorf("failed to get LLM review: %w", err)
	}

	// Validate and enhance review
	if err := s.validateReview(review, files); err != nil {
		log.Printf("Review validation warning: %v", err)
		// Continue with potentially corrected review
	}

	// Post review to GitHub
	if err := s.reviewPoster.PostReview(ctx, owner, repoName, prNumber, review, files); err != nil {
		return fmt.Errorf("failed to post review: %w", err)
	}

	log.Printf("Successfully reviewed PR #%d: %s", prNumber, review.Decision)
	return nil
}

// validateReview checks the review for common issues and filters invalid comments
func (s *Service) validateReview(review *types.ReviewResponse, files []*github.CommitFile) error {
	// Create file map for validation
	fileMap := make(map[string]*github.CommitFile)
	for _, file := range files {
		fileMap[file.GetFilename()] = file
	}

	// Filter out invalid file comments
	var validComments []types.FileComment
	var warnings []string

	for _, comment := range review.FileComments {
		// Check if file exists in PR
		file, exists := fileMap[comment.Path]
		if !exists {
			warnings = append(warnings, fmt.Sprintf("comment references non-existent file: %s", comment.Path))
			continue
		}

		// Check if line is in the diff
		if !gh.IsLineInDiff(file, comment.Line) {
			warnings = append(warnings, fmt.Sprintf("comment references line not in diff: %s:%d", comment.Path, comment.Line))
			continue
		}

		validComments = append(validComments, comment)
	}

	// Update review with valid comments only
	review.FileComments = validComments

	if len(warnings) > 0 {
		return fmt.Errorf("review validation warnings: %v", warnings)
	}

	return nil
}

// GetReviewStats returns statistics about the review
func (s *Service) GetReviewStats(review *types.ReviewResponse) ReviewStats {
	stats := ReviewStats{
		Decision:          review.Decision,
		GeneralComments:   len(review.GeneralComments),
		FileComments:      len(review.FileComments),
		HasBlockingIssues: review.HasBlockingIssues(),
	}

	// Count by severity
	for _, comment := range review.GeneralComments {
		switch comment.Severity {
		case types.SeverityError:
			stats.ErrorCount++
		case types.SeverityWarning:
			stats.WarningCount++
		case types.SeverityInfo:
			stats.InfoCount++
		}
	}

	for _, comment := range review.FileComments {
		switch comment.Severity {
		case types.SeverityError:
			stats.ErrorCount++
		case types.SeverityWarning:
			stats.WarningCount++
		case types.SeverityInfo:
			stats.InfoCount++
		}
	}

	// Count by type
	for _, comment := range review.FileComments {
		switch comment.Type {
		case types.TypeBug:
			stats.BugCount++
		case types.TypeSecurity:
			stats.SecurityCount++
		case types.TypePerformance:
			stats.PerformanceCount++
		case types.TypeStyle:
			stats.StyleCount++
		case types.TypeMaintainability:
			stats.MaintainabilityCount++
		}
	}

	return stats
}
