package llm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

func newEmbedTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestOllamaEmbedder_Embed(t *testing.T) {
	want := []float64{0.1, 0.2, 0.3}

	srv := newEmbedTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"embeddings": [][]float64{want},
		})
	})

	e := llm.NewOllamaEmbedder(srv.URL, "test-model", "", 5*time.Second)
	got, err := e.Embed(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("Embed() error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("Embed() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Embed()[%d] = %f, want %f", i, got[i], want[i])
		}
	}
}

func TestOllamaEmbedder_EmbedBatch(t *testing.T) {
	want := [][]float64{{0.1, 0.2}, {0.3, 0.4}}

	srv := newEmbedTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"embeddings": want,
		})
	})

	e := llm.NewOllamaEmbedder(srv.URL, "test-model", "", 5*time.Second)
	got, err := e.EmbedBatch(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatalf("EmbedBatch() error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("EmbedBatch() len = %d, want %d", len(got), len(want))
	}
}

func TestOllamaEmbedder_Embed_HTTPError(t *testing.T) {
	srv := newEmbedTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	e := llm.NewOllamaEmbedder(srv.URL, "test-model", "", 5*time.Second)
	_, err := e.Embed(context.Background(), "hello")
	if err == nil {
		t.Fatal("Embed() expected error for 503, got nil")
	}
}

func TestOllamaEmbedder_Embed_EmptyResponse(t *testing.T) {
	srv := newEmbedTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"embeddings": [][]float64{}})
	})

	e := llm.NewOllamaEmbedder(srv.URL, "test-model", "", 5*time.Second)
	_, err := e.Embed(context.Background(), "hello")
	if err == nil {
		t.Fatal("Embed() expected error for empty embeddings, got nil")
	}
}

func TestOllamaEmbedder_Ping_OK(t *testing.T) {
	srv := newEmbedTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/show" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})

	e := llm.NewOllamaEmbedder(srv.URL, "test-model", "", 5*time.Second)
	if err := e.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error: %v", err)
	}
}

func TestOllamaEmbedder_Ping_Error(t *testing.T) {
	srv := newEmbedTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	e := llm.NewOllamaEmbedder(srv.URL, "test-model", "", 5*time.Second)
	if err := e.Ping(context.Background()); err == nil {
		t.Fatal("Ping() expected error for 404, got nil")
	}
}

func TestOllamaEmbedder_APIKey(t *testing.T) {
	const wantKey = "secret-token"

	srv := newEmbedTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			return
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+wantKey {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer "+wantKey)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"embeddings": [][]float64{{0.1}},
		})
	})

	e := llm.NewOllamaEmbedder(srv.URL, "test-model", wantKey, 5*time.Second)
	if _, err := e.Embed(context.Background(), "hello"); err != nil {
		t.Fatalf("Embed() error: %v", err)
	}
}
