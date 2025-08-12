package reviewer

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v74/github"
	githubpkg "github.com/lehigh-university-libraries/mountain-hawk/internal/github"
)

// ContextBuilder builds context for LLM review
type ContextBuilder struct {
	githubClient *githubpkg.Client
}

// NewContextBuilder creates a new context builder
func NewContextBuilder(githubClient *githubpkg.Client) *ContextBuilder {
	return &ContextBuilder{
		githubClient: githubClient,
	}
}

// BuildContext creates a comprehensive context string for LLM review
func (cb *ContextBuilder) BuildContext(ctx context.Context, repo *github.Repository, pr *github.PullRequest, files []*github.CommitFile) (string, error) {
	var context strings.Builder

	// Add PR metadata
	cb.addPRMetadata(&context, pr)

	// Add file changes
	if err := cb.addFileChanges(ctx, &context, repo, pr, files); err != nil {
		return "", fmt.Errorf("failed to add file changes: %w", err)
	}

	return context.String(), nil
}

// addPRMetadata adds pull request metadata to context
func (cb *ContextBuilder) addPRMetadata(context *strings.Builder, pr *github.PullRequest) {
	context.WriteString(fmt.Sprintf("PR Title: %s\n", pr.GetTitle()))
	context.WriteString(fmt.Sprintf("PR Description: %s\n", pr.GetBody()))
	context.WriteString(fmt.Sprintf("Author: %s\n", pr.GetUser().GetLogin()))
	context.WriteString(fmt.Sprintf("Base Branch: %s\n", pr.GetBase().GetRef()))
	context.WriteString(fmt.Sprintf("Head Branch: %s\n", pr.GetHead().GetRef()))
	context.WriteString(fmt.Sprintf("Additions: %d, Deletions: %d\n\n", pr.GetAdditions(), pr.GetDeletions()))
}

// addFileChanges adds file content and changes to context
func (cb *ContextBuilder) addFileChanges(ctx context.Context, context *strings.Builder, repo *github.Repository, pr *github.PullRequest, files []*github.CommitFile) error {
	context.WriteString("Files changed:\n\n")

	owner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()
	headSHA := pr.GetHead().GetSHA()

	for _, file := range files {
		filename := file.GetFilename()
		status := file.GetStatus()

		// Skip removed files
		if status == "removed" {
			continue
		}

		// Skip binary files and large files
		if cb.shouldSkipFile(file) {
			context.WriteString(fmt.Sprintf("=== %s ===\n", filename))
			context.WriteString(fmt.Sprintf("Status: %s (skipped - binary or too large)\n\n", status))
			continue
		}

		// Get file content
		content, err := cb.githubClient.GetFileContent(ctx, owner, repoName, filename, headSHA)
		if err != nil {
			log.Printf("Error getting file %s: %v", filename, err)
			context.WriteString(fmt.Sprintf("=== %s ===\n", filename))
			context.WriteString(fmt.Sprintf("Status: %s (error reading file)\n\n", status))
			continue
		}

		// Add file information
		context.WriteString(fmt.Sprintf("=== %s ===\n", filename))
		context.WriteString(fmt.Sprintf("Status: %s\n", status))
		context.WriteString(fmt.Sprintf("Language: %s\n", cb.detectLanguage(filename)))
		context.WriteString(fmt.Sprintf("Additions: %d, Deletions: %d\n", file.GetAdditions(), file.GetDeletions()))

		// Add patch/diff if available
		if patch := file.GetPatch(); patch != "" {
			context.WriteString("Diff:\n")
			context.WriteString(patch)
			context.WriteString("\n")
		}

		// Add file content (truncated if too long)
		context.WriteString("Content:\n")
		truncatedContent := cb.truncateContent(content, 2000) // Limit content length
		context.WriteString(truncatedContent)
		context.WriteString("\n\n")
	}

	return nil
}

// shouldSkipFile determines if a file should be skipped from review
func (cb *ContextBuilder) shouldSkipFile(file *github.CommitFile) bool {
	filename := file.GetFilename()

	// Skip binary files
	if cb.isBinaryFile(filename) {
		return true
	}

	// Skip very large files
	if file.GetChanges() > 1000 {
		return true
	}

	// Skip generated files
	if cb.isGeneratedFile(filename) {
		return true
	}

	return false
}

// isBinaryFile checks if a file is likely binary
func (cb *ContextBuilder) isBinaryFile(filename string) bool {
	binaryExtensions := []string{
		".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".svg",
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".zip", ".tar", ".gz", ".7z", ".rar",
		".exe", ".dll", ".so", ".dylib",
		".woff", ".woff2", ".ttf", ".eot",
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, binaryExt := range binaryExtensions {
		if ext == binaryExt {
			return true
		}
	}

	return false
}

// isGeneratedFile checks if a file is generated/vendor code
func (cb *ContextBuilder) isGeneratedFile(filename string) bool {
	generatedPatterns := []string{
		"vendor/", "node_modules/", ".git/",
		"dist/", "build/", "target/",
		".min.js", ".min.css",
		"package-lock.json", "yarn.lock", "Cargo.lock",
		".pb.go", "_gen.go", ".generated.",
	}

	lowerFilename := strings.ToLower(filename)
	for _, pattern := range generatedPatterns {
		if strings.Contains(lowerFilename, pattern) {
			return true
		}
	}

	return false
}

// detectLanguage attempts to detect the programming language
func (cb *ContextBuilder) detectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	languageMap := map[string]string{
		".go":    "Go",
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".py":    "Python",
		".java":  "Java",
		".cpp":   "C++",
		".c":     "C",
		".cs":    "C#",
		".rb":    "Ruby",
		".php":   "PHP",
		".rs":    "Rust",
		".kt":    "Kotlin",
		".swift": "Swift",
		".scala": "Scala",
		".sh":    "Shell",
		".sql":   "SQL",
		".html":  "HTML",
		".css":   "CSS",
		".scss":  "SCSS",
		".yaml":  "YAML",
		".yml":   "YAML",
		".json":  "JSON",
		".xml":   "XML",
		".md":    "Markdown",
	}

	if lang, exists := languageMap[ext]; exists {
		return lang
	}

	// Check for specific filenames
	basename := strings.ToLower(filepath.Base(filename))
	switch basename {
	case "dockerfile":
		return "Docker"
	case "makefile":
		return "Makefile"
	case "rakefile":
		return "Ruby"
	case "gemfile":
		return "Ruby"
	}

	return "Unknown"
}

// truncateContent truncates content to a maximum length
func (cb *ContextBuilder) truncateContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}

	truncated := content[:maxLength]
	return truncated + "\n... (content truncated)"
}
