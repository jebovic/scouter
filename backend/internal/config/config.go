package config

import (
	"fmt"
	"os"
)

// Config holds all environment-derived configuration for the server.
type Config struct {
	DatabaseURL      string
	AnthropicAPIKey  string
	LLMProvider      string // "anthropic" | "ollama"
	OllamaBaseURL    string
	OllamaModel      string
	Port             string
	Env              string // "development" | "production"
}

// Load reads required environment variables, returning an error if any are missing.
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		LLMProvider:     os.Getenv("LLM_PROVIDER"),
		OllamaBaseURL:   os.Getenv("OLLAMA_BASE_URL"),
		OllamaModel:     os.Getenv("OLLAMA_MODEL"),
		Port:            os.Getenv("PORT"),
		Env:             os.Getenv("ENV"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.LLMProvider == "" {
		cfg.LLMProvider = "ollama"
	}
	if cfg.OllamaBaseURL == "" {
		cfg.OllamaBaseURL = "http://host.docker.internal:11434"
	}
	if cfg.OllamaModel == "" {
		cfg.OllamaModel = "qwen2.5:7b"
	}
	// Anthropic key is required only when used as primary or fallback.
	// In routing mode both are needed; in ollama-only mode it is optional.
	if cfg.LLMProvider == "anthropic" && cfg.AnthropicAPIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required when LLM_PROVIDER=anthropic")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.Env == "" {
		cfg.Env = "production"
	}

	return cfg, nil
}
