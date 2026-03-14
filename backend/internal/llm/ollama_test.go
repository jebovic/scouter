package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNewOllamaProvider_defaults verifies that a provider created with no options
// has the 180s default timeout and an empty apiKey.
func TestNewOllamaProvider_defaults(t *testing.T) {
	p := NewOllamaProvider("http://localhost:11434", "qwen3:14b")

	if p.timeout != 180*time.Second {
		t.Errorf("default timeout: want 180s, got %v", p.timeout)
	}
	if p.apiKey != "" {
		t.Errorf("default apiKey: want empty, got %q", p.apiKey)
	}
	if p.client == nil {
		t.Fatal("client must not be nil")
	}
	if p.client.Timeout != 180*time.Second {
		t.Errorf("client.Timeout: want 180s, got %v", p.client.Timeout)
	}
}

// TestNewOllamaProvider_withTimeout verifies that WithTimeout overrides the client timeout.
func TestNewOllamaProvider_withTimeout(t *testing.T) {
	p := NewOllamaProvider("http://localhost:11434", "qwen3:4b", WithTimeout(30*time.Second))

	if p.timeout != 30*time.Second {
		t.Errorf("timeout: want 30s, got %v", p.timeout)
	}
	if p.client.Timeout != 30*time.Second {
		t.Errorf("client.Timeout: want 30s, got %v", p.client.Timeout)
	}
}

// TestNewOllamaProvider_withAPIKey verifies that WithAPIKey sets the apiKey field.
func TestNewOllamaProvider_withAPIKey(t *testing.T) {
	p := NewOllamaProvider("http://localhost:11434", "qwen3:14b", WithAPIKey("secret-key"))

	if p.apiKey != "secret-key" {
		t.Errorf("apiKey: want %q, got %q", "secret-key", p.apiKey)
	}
}

// TestOllamaComplete_setsAuthHeader verifies that when an apiKey is set,
// the Authorization: Bearer <key> header is sent on every request.
func TestOllamaComplete_setsAuthHeader(t *testing.T) {
	var capturedAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		resp := ollamaChatResponse{
			Message: ollamaMessage{Role: "assistant", Content: "hello"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := NewOllamaProvider(srv.URL, "qwen3:14b", WithAPIKey("test-token"))
	_, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	want := "Bearer test-token"
	if capturedAuth != want {
		t.Errorf("Authorization header: want %q, got %q", want, capturedAuth)
	}
}

// TestOllamaComplete_noAuthHeader verifies that without an apiKey,
// no Authorization header is sent.
func TestOllamaComplete_noAuthHeader(t *testing.T) {
	var capturedAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		resp := ollamaChatResponse{
			Message: ollamaMessage{Role: "assistant", Content: "hello"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := NewOllamaProvider(srv.URL, "qwen3:14b")
	_, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	if capturedAuth != "" {
		t.Errorf("Authorization header: want empty, got %q", capturedAuth)
	}
}

// TestOllamaComplete_setsModelName verifies that the response ModelName equals
// the model string passed to NewOllamaProvider.
func TestOllamaComplete_setsModelName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ollamaChatResponse{
			Message: ollamaMessage{Role: "assistant", Content: "pong"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	const modelName = "qwen3:14b"
	p := NewOllamaProvider(srv.URL, modelName)
	got, err := p.Complete(context.Background(), CompletionRequest{
		Messages: []Message{{Role: "user", Content: "ping"}},
	})
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	if got.ModelName != modelName {
		t.Errorf("ModelName: want %q, got %q", modelName, got.ModelName)
	}
}

// TestOllama_Ping_success verifies that Ping returns nil when the server responds 200.
func TestOllama_Ping_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Ping method: want POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/show" {
			t.Errorf("Ping path: want /api/show, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := NewOllamaProvider(srv.URL, "qwen3:14b")
	if err := p.Ping(context.Background()); err != nil {
		t.Errorf("Ping success: want nil error, got %v", err)
	}
}

// TestOllama_Ping_failure verifies that Ping returns an error when the server responds non-200.
func TestOllama_Ping_failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := NewOllamaProvider(srv.URL, "qwen3:14b")
	if err := p.Ping(context.Background()); err == nil {
		t.Error("Ping failure: want error on 500, got nil")
	}
}

// TestOllama_Ping_networkError verifies that Ping returns an error on network failure.
func TestOllama_Ping_networkError(t *testing.T) {
	// Use an invalid URL that will never connect.
	p := NewOllamaProvider("http://127.0.0.1:1", "qwen3:14b")
	if err := p.Ping(context.Background()); err == nil {
		t.Error("Ping network error: want error on connection refused, got nil")
	}
}
