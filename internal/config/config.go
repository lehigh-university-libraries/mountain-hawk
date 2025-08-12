package config

import (
	"fmt"
	"log"
	"os"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Port string

	// GitHub configuration
	GitHubToken   string
	WebhookSecret string

	// LLM configuration
	OllamaURL   string
	OllamaModel string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnvOrDefault("PORT", "8080"),
		OllamaModel: getEnvOrDefault("OLLAMA_MODEL", "gpt-oss:20b"),
	}

	// Required environment variables
	required := map[string]*string{
		"GITHUB_TOKEN":   &cfg.GitHubToken,
		"OLLAMA_HOST":    &cfg.OllamaURL,
		"WEBHOOK_SECRET": &cfg.WebhookSecret,
	}

	var missing []string
	for envVar, field := range required {
		value := os.Getenv(envVar)
		if value == "" {
			missing = append(missing, envVar)
		} else {
			*field = value
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	return cfg, nil
}

// MustLoad loads configuration and panics on error
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return cfg
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
