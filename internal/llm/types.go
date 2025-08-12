package llm

// ProviderType represents different LLM providers
type ProviderType string

const (
	ProviderOllama    ProviderType = "ollama"
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
)

// Config holds LLM client configuration
type Config struct {
	Provider ProviderType
	BaseURL  string
	Model    string
	APIKey   string
	Timeout  int // seconds
}

// ModelCapabilities describes what a model can do
type ModelCapabilities struct {
	SupportsCodeReview bool
	MaxTokens          int
	ContextWindow      int
}

// Usage tracks token/request usage
type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	RequestCount int
}
