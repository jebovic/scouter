package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
)

// Config holds all environment-derived configuration for the server.
type Config struct {
	DatabaseURL          string
	AnthropicAPIKey      string
	LLMProvider          string // "anthropic" | "ollama"
	OllamaBaseURL        string
	OllamaModel          string // legacy alias; use OllamaHeavyModel for routing
	Port                 string
	Env                  string // "development" | "production"
	PriceCheckEnabled    bool
	PriceCheckCron       string // cron expression, e.g. "0 */6 * * *"
	PriceCheckMaxMissions int

	// Phase 9: Ollama Smart Routing
	OllamaHeavyModel   string
	OllamaFastModel    string
	OllamaHeavyTimeout time.Duration
	OllamaFastTimeout  time.Duration
	OllamaCloudURL     string
	OllamaCloudModel   string
	OllamaCloudAPIKey  string
	OllamaCloudRPM     int
	OllamaEmbedModel   string

	// Phase 14: Observability
	MetricsEnabled bool

	// Phase 29: Travel APIs
	AviationStackAPIKey  string
	AviationStackBaseURL string // default: "https://api.aviationstack.com"
	SNCFAPIKey           string
	SNCFBaseURL          string // default: "https://api.sncf.com/v1"

	// Translation worker
	SupportedLocales string // comma-separated, e.g. "fr,de"

	// Product Images: MinIO object storage
	MinioEndpoint  string
	MinioPublicURL string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioQuotaMB   int64
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
	// Backward compat: keep OllamaModel default as "qwen2.5:7b".
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

	cfg.PriceCheckEnabled = os.Getenv("PRICE_CHECK_ENABLED") == "true"
	cfg.PriceCheckCron = os.Getenv("PRICE_CHECK_CRON")
	if cfg.PriceCheckCron == "" {
		cfg.PriceCheckCron = "0 */6 * * *"
	}
	if cfg.PriceCheckEnabled {
		if _, err := cron.ParseStandard(cfg.PriceCheckCron); err != nil {
			return nil, fmt.Errorf("PRICE_CHECK_CRON %q is not a valid cron expression: %w", cfg.PriceCheckCron, err)
		}
	}
	if v := os.Getenv("PRICE_CHECK_MAX_MISSIONS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.PriceCheckMaxMissions = n
		}
	}
	if cfg.PriceCheckMaxMissions == 0 {
		cfg.PriceCheckMaxMissions = 10
	}

	// ── Phase 9: Ollama Smart Routing ─────────────────────────────────────────

	// OllamaHeavyModel: OLLAMA_HEAVY_MODEL > OLLAMA_MODEL > default.
	cfg.OllamaHeavyModel = os.Getenv("OLLAMA_HEAVY_MODEL")
	if cfg.OllamaHeavyModel == "" {
		if legacyModel := os.Getenv("OLLAMA_MODEL"); legacyModel != "" {
			cfg.OllamaHeavyModel = legacyModel
		} else {
			cfg.OllamaHeavyModel = "qwen3:14b"
		}
	}

	// OllamaFastModel: OLLAMA_FAST_MODEL > default.
	cfg.OllamaFastModel = os.Getenv("OLLAMA_FAST_MODEL")
	if cfg.OllamaFastModel == "" {
		cfg.OllamaFastModel = "qwen3:4b"
	}

	// OllamaHeavyTimeout: parse OLLAMA_HEAVY_TIMEOUT seconds, default 180s.
	cfg.OllamaHeavyTimeout = parseDurationSeconds("OLLAMA_HEAVY_TIMEOUT", 180)

	// OllamaFastTimeout: parse OLLAMA_FAST_TIMEOUT seconds, default 60s.
	cfg.OllamaFastTimeout = parseDurationSeconds("OLLAMA_FAST_TIMEOUT", 60)

	// Cloud fields — empty string disables cloud routing.
	cfg.OllamaCloudURL = os.Getenv("OLLAMA_CLOUD_URL")
	cfg.OllamaCloudModel = os.Getenv("OLLAMA_CLOUD_MODEL")
	cfg.OllamaCloudAPIKey = os.Getenv("OLLAMA_CLOUD_API_KEY")

	// OllamaCloudRPM: default 10.
	cfg.OllamaCloudRPM = 10
	if v := os.Getenv("OLLAMA_CLOUD_RPM"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.OllamaCloudRPM = n
		}
	}

	// OllamaEmbedModel: default "mxbai-embed-large".
	cfg.OllamaEmbedModel = os.Getenv("OLLAMA_EMBED_MODEL")
	if cfg.OllamaEmbedModel == "" {
		cfg.OllamaEmbedModel = "mxbai-embed-large"
	}

	// Phase 14: Metrics
	cfg.MetricsEnabled = os.Getenv("METRICS_ENABLED") == "true"

	// Phase 29: Travel APIs
	cfg.AviationStackAPIKey = os.Getenv("AVIATIONSTACK_API_KEY")
	cfg.AviationStackBaseURL = os.Getenv("AVIATIONSTACK_BASE_URL")
	if cfg.AviationStackBaseURL == "" {
		cfg.AviationStackBaseURL = "https://api.aviationstack.com"
	}
	cfg.SNCFAPIKey = os.Getenv("SNCF_API_KEY")
	cfg.SNCFBaseURL = os.Getenv("SNCF_BASE_URL")
	if cfg.SNCFBaseURL == "" {
		cfg.SNCFBaseURL = "https://api.sncf.com/v1"
	}

	// Translation worker
	cfg.SupportedLocales = os.Getenv("SUPPORTED_LOCALES")
	if cfg.SupportedLocales == "" {
		cfg.SupportedLocales = "fr,de"
	}

	// Product Images: MinIO object storage
	cfg.MinioEndpoint = os.Getenv("MINIO_ENDPOINT")
	if cfg.MinioEndpoint == "" {
		cfg.MinioEndpoint = "minio:9000"
	}
	cfg.MinioPublicURL = os.Getenv("MINIO_PUBLIC_URL")
	if cfg.MinioPublicURL == "" {
		cfg.MinioPublicURL = "https://minio.dev.local"
	}
	cfg.MinioAccessKey = os.Getenv("MINIO_ACCESS_KEY")
	cfg.MinioSecretKey = os.Getenv("MINIO_SECRET_KEY")
	cfg.MinioBucket = os.Getenv("MINIO_BUCKET")
	if cfg.MinioBucket == "" {
		cfg.MinioBucket = "scouter-images"
	}
	cfg.MinioQuotaMB = 2048
	if v := os.Getenv("MINIO_QUOTA_MB"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			cfg.MinioQuotaMB = n
		}
	}

	return cfg, nil
}

// parseDurationSeconds reads an env var as integer seconds and returns a
// time.Duration. Falls back to defaultSec when the variable is absent or invalid.
func parseDurationSeconds(envKey string, defaultSec int) time.Duration {
	if v := os.Getenv(envKey); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return time.Duration(defaultSec) * time.Second
}
