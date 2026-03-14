package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// OllamaEmbedder implements EmbedProvider using the Ollama /api/embed endpoint.
type OllamaEmbedder struct {
	baseURL string
	model   string
	apiKey  string
	client  *http.Client
}

// NewOllamaEmbedder creates an OllamaEmbedder with the given parameters.
func NewOllamaEmbedder(baseURL, model, apiKey string, timeout time.Duration) *OllamaEmbedder {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &OllamaEmbedder{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: timeout},
	}
}

// ollamaEmbedRequest is the request body for POST /api/embed.
type ollamaEmbedRequest struct {
	Model string `json:"model"`
	Input any    `json:"input"` // string or []string
}

// ollamaEmbedResponse is the response body from POST /api/embed.
type ollamaEmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

// Embed returns a single embedding vector for the given text.
func (e *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	vecs, err := e.embedBatch(ctx, text)
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("ollama embed: empty embeddings in response")
	}
	return vecs[0], nil
}

// EmbedBatch returns embedding vectors for multiple texts in a single call.
func (e *OllamaEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	return e.embedBatch(ctx, texts)
}

// embedBatch calls /api/embed with either a string or []string input.
func (e *OllamaEmbedder) embedBatch(ctx context.Context, input any) ([][]float64, error) {
	body, err := json.Marshal(ollamaEmbedRequest{Model: e.model, Input: input})
	if err != nil {
		return nil, fmt.Errorf("ollama embed: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama embed: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama embed: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &StatusError{Code: resp.StatusCode}
	}

	var embedResp ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("ollama embed: decode response: %w", err)
	}
	if len(embedResp.Embeddings) == 0 {
		return nil, fmt.Errorf("ollama embed: no embeddings returned")
	}
	return embedResp.Embeddings, nil
}

// Ping checks liveness by calling POST /api/show with the embed model name.
func (e *OllamaEmbedder) Ping(ctx context.Context) error {
	body, err := json.Marshal(map[string]string{"name": e.model})
	if err != nil {
		return fmt.Errorf("embed ping: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/api/show", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("embed ping: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("embed ping: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("embed ping: unexpected status %d", resp.StatusCode)
	}
	return nil
}
