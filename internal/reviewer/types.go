package reviewer

import "github.com/lehigh-university-libraries/mountain-hawk/pkg/types"

// ReviewStats contains statistics about a review
type ReviewStats struct {
	Decision          types.ReviewDecision
	GeneralComments   int
	FileComments      int
	HasBlockingIssues bool

	// Count by severity
	ErrorCount   int
	WarningCount int
	InfoCount    int

	// Count by type
	BugCount             int
	SecurityCount        int
	PerformanceCount     int
	StyleCount           int
	MaintainabilityCount int
}

// ReviewOptions contains options for customizing review behavior
type ReviewOptions struct {
	// Focus areas
	FocusOnSecurity    bool
	FocusOnPerformance bool
	FocusOnStyle       bool
	SkipTests          bool
	SkipGenerated      bool

	// LLM options
	Model       string
	MaxTokens   int
	Temperature float32

	// Review behavior
	AutoApproveSimple       bool
	RequireExplicitApproval bool
	MaxCommentsPerFile      int
}

// ReviewContext contains additional context for review
type ReviewContext struct {
	Repository string
	Branch     string
	Author     string
	IsDraft    bool
	IsHotfix   bool
	Labels     []string
	Reviewers  []string
}

// ReviewResult contains the complete result of a review operation
type ReviewResult struct {
	Success  bool
	Review   *types.ReviewResponse
	Stats    ReviewStats
	Errors   []error
	Warnings []string
	Duration int64 // milliseconds
}
