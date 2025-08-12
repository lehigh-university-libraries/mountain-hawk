package github

import (
	"context"

	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client with our application-specific methods
type Client struct {
	client *github.Client
}

// NewClient creates a new GitHub client with authentication
func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
	}
}

// GetPRFiles retrieves all files changed in a pull request
func (c *Client) GetPRFiles(ctx context.Context, owner, repo string, prNumber int) ([]*github.CommitFile, error) {
	files, _, err := c.client.PullRequests.ListFiles(ctx, owner, repo, prNumber, nil)
	return files, err
}

// GetFileContent retrieves the content of a file at a specific commit
func (c *Client) GetFileContent(ctx context.Context, owner, repo, path, ref string) (string, error) {
	fileContent, _, _, err := c.client.Repositories.GetContents(
		ctx, owner, repo, path,
		&github.RepositoryContentGetOptions{Ref: ref},
	)
	if err != nil {
		return "", err
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", err
	}

	return content, nil
}

// CreateReview creates a pull request review with comments
func (c *Client) CreateReview(ctx context.Context, owner, repo string, prNumber int, review *github.PullRequestReviewRequest) error {
	_, _, err := c.client.PullRequests.CreateReview(ctx, owner, repo, prNumber, review)
	return err
}

// CreateIssueComment creates a general comment on the pull request
func (c *Client) CreateIssueComment(ctx context.Context, owner, repo string, prNumber int, body string) error {
	comment := &github.IssueComment{
		Body: &body,
	}
	_, _, err := c.client.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
	return err
}

// GetPullRequest retrieves a pull request and repository
func (c *Client) GetPullRequest(ctx context.Context, owner, repo string, prNumber int) (*github.PullRequest, *github.Repository, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, nil, err
	}

	repository, _, err := c.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, nil, err
	}

	return pr, repository, nil
}
