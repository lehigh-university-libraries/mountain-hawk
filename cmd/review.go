// cmd/review.go
package cmd

import (
	"context"
	"fmt"

	"github.com/lehigh-university-libraries/mountain-hawk/internal/config"
	"github.com/lehigh-university-libraries/mountain-hawk/internal/github"
	"github.com/lehigh-university-libraries/mountain-hawk/internal/llm"
	"github.com/lehigh-university-libraries/mountain-hawk/internal/reviewer"
	"github.com/spf13/cobra"
)

var (
	// Review command flags
	owner string
	repo  string
	pr    int
)

// NewReviewCommand creates the review command
func NewReviewCommand() *cobra.Command {
	reviewCmd := &cobra.Command{
		Use:   "review",
		Short: "Review a specific pull request",
		Long: `Review a specific pull request by providing the repository owner, name, and PR number.
This will fetch the PR data via MCP, analyze it with AI, and provide structured feedback.`,
		Example: `  # Review a specific PR
  mountain-hawk review --owner=microsoft --repo=vscode --pr=123456
  
  # Review with verbose output
  mountain-hawk review --owner=facebook --repo=react --pr=5678 --verbose`,
		RunE: runReview,
	}

	// Review command flags
	reviewCmd.Flags().StringVarP(&owner, "owner", "o", "", "Repository owner (required)")
	reviewCmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository name (required)")
	reviewCmd.Flags().IntVarP(&pr, "pr", "n", 0, "Pull request number (required)")
	reviewCmd.MarkFlagRequired("owner")
	reviewCmd.MarkFlagRequired("repo")
	reviewCmd.MarkFlagRequired("pr")

	return reviewCmd
}

func runReview(cmd *cobra.Command, args []string) error {
	cfg := config.MustLoad()
	verbose := GetVerbose()
	if verbose {
		fmt.Printf("Reviewing PR #%d in %s/%s...\n", pr, owner, repo)
		fmt.Printf("LLM: %s (%s)\n", cfg.OllamaURL, cfg.OllamaModel)
	}

	// Initialize services
	githubClient := github.NewClient(cfg.GitHubToken)
	llmClient := llm.NewOllamaClient(cfg.OllamaURL, cfg.OllamaModel)
	reviewService := reviewer.NewService(githubClient, llmClient)

	if verbose {
		fmt.Println("Fetching PR details...")
	}

	// Get PR details
	prData, repository, err := githubClient.GetPullRequest(context.Background(), owner, repo, pr)
	if err != nil {
		return fmt.Errorf("failed to get PR: %w", err)
	}

	if verbose {
		fmt.Printf("Found PR: %s\n", prData.GetTitle())
		fmt.Printf("Author: %s\n", prData.GetUser().GetLogin())
		fmt.Printf("Changed files: %d\n", prData.GetChangedFiles())
		fmt.Printf("Additions: %d, Deletions: %d\n", prData.GetAdditions(), prData.GetDeletions())
		fmt.Println("Starting review process...")
	}

	// Review the PR
	err = reviewService.ReviewPR(prData, repository)
	if err != nil {
		return fmt.Errorf("failed to review PR: %w", err)
	}

	return nil
}
