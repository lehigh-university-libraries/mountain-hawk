package types

// ReviewDecision represents the possible review decisions
type ReviewDecision string

const (
	DecisionApprove        ReviewDecision = "approve"
	DecisionRequestChanges ReviewDecision = "request_changes"
	DecisionComment        ReviewDecision = "comment"
)

// Severity levels for comments
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// CommentType categorizes the type of feedback
type CommentType string

const (
	TypeBug             CommentType = "bug"
	TypeStyle           CommentType = "style"
	TypePerformance     CommentType = "performance"
	TypeSecurity        CommentType = "security"
	TypeMaintainability CommentType = "maintainability"
)

// ReviewResponse represents the structured response from the LLM
type ReviewResponse struct {
	Decision          ReviewDecision   `json:"decision"`
	DecisionRationale string           `json:"decision_rationale"`
	GeneralComments   []GeneralComment `json:"general_comments"`
	FileComments      []FileComment    `json:"file_comments"`
	Summary           string           `json:"summary,omitempty"`
}

// GeneralComment represents overall feedback about the PR
type GeneralComment struct {
	Body     string   `json:"body"`
	Severity Severity `json:"severity"`
}

// FileComment represents line-specific feedback
type FileComment struct {
	Path     string      `json:"path"`
	Line     int         `json:"line"`
	Body     string      `json:"body"`
	Severity Severity    `json:"severity"`
	Type     CommentType `json:"type"`
}

// IsBlockingDecision returns true if the decision blocks the PR
func (d ReviewDecision) IsBlockingDecision() bool {
	return d == DecisionRequestChanges
}

// IsError returns true if the severity indicates an error
func (s Severity) IsError() bool {
	return s == SeverityError
}

// IsSecurityIssue returns true if the comment type is security-related
func (t CommentType) IsSecurityIssue() bool {
	return t == TypeSecurity
}

// HasBlockingIssues returns true if the review contains any blocking issues
func (r *ReviewResponse) HasBlockingIssues() bool {
	if r.Decision.IsBlockingDecision() {
		return true
	}

	// Check for error-level comments
	for _, comment := range r.GeneralComments {
		if comment.Severity.IsError() {
			return true
		}
	}

	for _, comment := range r.FileComments {
		if comment.Severity.IsError() {
			return true
		}
	}

	return false
}

// GetSecurityIssues returns all security-related comments
func (r *ReviewResponse) GetSecurityIssues() []FileComment {
	var securityIssues []FileComment
	for _, comment := range r.FileComments {
		if comment.Type.IsSecurityIssue() {
			securityIssues = append(securityIssues, comment)
		}
	}
	return securityIssues
}

// GetErrorComments returns all error-level comments
func (r *ReviewResponse) GetErrorComments() []FileComment {
	var errorComments []FileComment
	for _, comment := range r.FileComments {
		if comment.Severity.IsError() {
			errorComments = append(errorComments, comment)
		}
	}
	return errorComments
}
