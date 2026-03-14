package config

import (
	"testing"
	"time"
)

// TestLoad_OllamaHeavyModel_fromOllamaModel verifies backward compat:
// when only OLLAMA_MODEL is set, OllamaHeavyModel takes that value.
func TestLoad_OllamaHeavyModel_fromOllamaModel(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("OLLAMA_MODEL", "custom:7b")
	t.Setenv("OLLAMA_HEAVY_MODEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.OllamaHeavyModel != "custom:7b" {
		t.Errorf("OllamaHeavyModel: want %q, got %q", "custom:7b", cfg.OllamaHeavyModel)
	}
}

// TestLoad_OllamaHeavyModel_fromHeavyModelEnv verifies that OLLAMA_HEAVY_MODEL
// is used when set, and OllamaModel falls back to "qwen2.5:7b".
func TestLoad_OllamaHeavyModel_fromHeavyModelEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("OLLAMA_MODEL", "")
	t.Setenv("OLLAMA_HEAVY_MODEL", "qwen3:14b")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.OllamaHeavyModel != "qwen3:14b" {
		t.Errorf("OllamaHeavyModel: want %q, got %q", "qwen3:14b", cfg.OllamaHeavyModel)
	}
	// OllamaModel legacy field should still have a default
	if cfg.OllamaModel != "qwen2.5:7b" {
		t.Errorf("OllamaModel legacy default: want %q, got %q", "qwen2.5:7b", cfg.OllamaModel)
	}
}

// TestLoad_OllamaHeavyModel_heavyWinsOverLegacy verifies that when both are set,
// OLLAMA_HEAVY_MODEL takes precedence.
func TestLoad_OllamaHeavyModel_heavyWinsOverLegacy(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("OLLAMA_MODEL", "legacy:7b")
	t.Setenv("OLLAMA_HEAVY_MODEL", "qwen3:14b")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.OllamaHeavyModel != "qwen3:14b" {
		t.Errorf("OllamaHeavyModel: want %q (HEAVY wins), got %q", "qwen3:14b", cfg.OllamaHeavyModel)
	}
}

// TestLoad_OllamaHeavyModel_neitherSet verifies that when no model envs are set,
// OllamaHeavyModel defaults to "qwen3:14b".
func TestLoad_OllamaHeavyModel_neitherSet(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("OLLAMA_MODEL", "")
	t.Setenv("OLLAMA_HEAVY_MODEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.OllamaHeavyModel != "qwen3:14b" {
		t.Errorf("OllamaHeavyModel default: want %q, got %q", "qwen3:14b", cfg.OllamaHeavyModel)
	}
}

// TestLoad_OllamaCloudURL_emptyDisablesCloud verifies that an empty OLLAMA_CLOUD_URL
// results in an empty OllamaCloudURL (cloud disabled).
func TestLoad_OllamaCloudURL_emptyDisablesCloud(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("OLLAMA_CLOUD_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.OllamaCloudURL != "" {
		t.Errorf("OllamaCloudURL: want empty (cloud disabled), got %q", cfg.OllamaCloudURL)
	}
}

// TestLoad_OllamaHeavyTimeout_parsing verifies that OLLAMA_HEAVY_TIMEOUT=120
// results in a 120-second duration.
func TestLoad_OllamaHeavyTimeout_parsing(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("OLLAMA_HEAVY_TIMEOUT", "120")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	want := 120 * time.Second
	if cfg.OllamaHeavyTimeout != want {
		t.Errorf("OllamaHeavyTimeout: want %v, got %v", want, cfg.OllamaHeavyTimeout)
	}
}

// TestLoad_OllamaFastTimeout_default verifies that OLLAMA_FAST_TIMEOUT defaults to 60s.
func TestLoad_OllamaFastTimeout_default(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("OLLAMA_FAST_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	want := 60 * time.Second
	if cfg.OllamaFastTimeout != want {
		t.Errorf("OllamaFastTimeout default: want %v, got %v", want, cfg.OllamaFastTimeout)
	}
}

// TestLoad_OllamaCloudRPM_default verifies that OLLAMA_CLOUD_RPM defaults to 10.
func TestLoad_OllamaCloudRPM_default(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("OLLAMA_CLOUD_RPM", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.OllamaCloudRPM != 10 {
		t.Errorf("OllamaCloudRPM default: want 10, got %d", cfg.OllamaCloudRPM)
	}
}

// TestLoad_OllamaFastModel_default verifies that OLLAMA_FAST_MODEL defaults to "qwen3:4b".
func TestLoad_OllamaFastModel_default(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("OLLAMA_FAST_MODEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.OllamaFastModel != "qwen3:4b" {
		t.Errorf("OllamaFastModel default: want %q, got %q", "qwen3:4b", cfg.OllamaFastModel)
	}
}

// TestLoad_OllamaEmbedModel_default verifies that OLLAMA_EMBED_MODEL defaults to "mxbai-embed-large".
func TestLoad_OllamaEmbedModel_default(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("OLLAMA_EMBED_MODEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.OllamaEmbedModel != "mxbai-embed-large" {
		t.Errorf("OllamaEmbedModel default: want %q, got %q", "mxbai-embed-large", cfg.OllamaEmbedModel)
	}
}

// TestLoad_AllPhase9Fields verifies all phase 9 fields load correctly when set.
func TestLoad_AllPhase9Fields(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("OLLAMA_HEAVY_MODEL", "qwen3:14b")
	t.Setenv("OLLAMA_FAST_MODEL", "qwen3:4b")
	t.Setenv("OLLAMA_HEAVY_TIMEOUT", "180")
	t.Setenv("OLLAMA_FAST_TIMEOUT", "45")
	t.Setenv("OLLAMA_CLOUD_URL", "https://ollama.com")
	t.Setenv("OLLAMA_CLOUD_MODEL", "deepseek-v3:cloud")
	t.Setenv("OLLAMA_CLOUD_API_KEY", "cloud-key-123")
	t.Setenv("OLLAMA_CLOUD_RPM", "20")
	t.Setenv("OLLAMA_EMBED_MODEL", "mxbai-embed-large")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	checks := []struct {
		name string
		got  any
		want any
	}{
		{"OllamaHeavyModel", cfg.OllamaHeavyModel, "qwen3:14b"},
		{"OllamaFastModel", cfg.OllamaFastModel, "qwen3:4b"},
		{"OllamaHeavyTimeout", cfg.OllamaHeavyTimeout, 180 * time.Second},
		{"OllamaFastTimeout", cfg.OllamaFastTimeout, 45 * time.Second},
		{"OllamaCloudURL", cfg.OllamaCloudURL, "https://ollama.com"},
		{"OllamaCloudModel", cfg.OllamaCloudModel, "deepseek-v3:cloud"},
		{"OllamaCloudAPIKey", cfg.OllamaCloudAPIKey, "cloud-key-123"},
		{"OllamaCloudRPM", cfg.OllamaCloudRPM, 20},
		{"OllamaEmbedModel", cfg.OllamaEmbedModel, "mxbai-embed-large"},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: want %v, got %v", c.name, c.want, c.got)
		}
	}
}
