# LLM Routing

## Overview

![LLM Router](../assets/llm-router.svg)

SCOUTER's SmartRouter (Phase 9) provides a capability-matched, fault-tolerant LLM routing layer. It routes agent requests to the best available provider, cascading on infrastructure failures.

---

## Provider Interface

```go
// internal/llm/provider.go
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
}
```

This interface is **transport-only**. Agents own all prompt engineering and response parsing.

---

## SmartRouter

### Cascade Order

```
1. Ollama Heavy (qwen3:14b)  — local, primary
2. Ollama Fast  (qwen3:4b)   — local, lightweight fallback
3. Ollama Cloud              — remote (deepseek-v3.2:cloud via ollama.com)
4. Anthropic                 — claude-sonnet-4-6, final fallback
```

**Cascade triggers on infrastructure errors only:**
- Connection timeout / refused
- Circuit breaker open
- Rate limit exceeded (own rate limit, not provider's)

**Does NOT cascade on:** bad LLM output — uses `RetryAsJSON` instead.

### Capability Matching

```go
type RequestOpts struct {
    Capability string  // "tool-use", "embedding", "fast", "heavy"
}

func WithRequestOpts(ctx context.Context, opts RequestOpts) context.Context
func HasRequestOpts(ctx context.Context) (RequestOpts, bool)
```

Agents hint at their needs:

```go
// ResearchAgent — needs tool use
ctx = llm.WithRequestOpts(ctx, llm.RequestOpts{Capability: "tool-use"})
resp, err := router.Complete(ctx, req)

// Fast classification agent — lightweight
ctx = llm.WithRequestOpts(ctx, llm.RequestOpts{Capability: "fast"})
```

SmartRouter selects providers that support the requested capability.

### Circuit Breakers

Each provider has a circuit breaker:

```
State: Closed → Open (after N consecutive failures) → Half-Open → Closed
```

When a circuit is open, SmartRouter skips that provider immediately (no latency hit) and tries the next.

### Rate Limiters

Per-model token bucket rate limiters prevent hammering a single provider. When a limiter is full, SmartRouter cascades to the next available provider.

### RetryAsJSON

When a provider returns non-JSON output for a JSON-expected call, SmartRouter retries with an explicit JSON formatting instruction:

```go
func RetryAsJSON(ctx context.Context, p Provider, req CompletionRequest) (CompletionResponse, error) {
    req.Messages = append(req.Messages, Message{
        Role: "user",
        Content: "Please respond with valid JSON only.",
    })
    return p.Complete(ctx, req)
}
```

---

## ModelPool

```go
type ModelPool struct {
    providers []ProviderWithCapabilities
}

func (p *ModelPool) ForCapabilities(caps ...string) []Provider {
    // Filter providers that support all requested capabilities
    // Return ordered by priority
}
```

---

## Anthropic Provider

```go
type AnthropicProvider struct {
    client *anthropic.Client
    model  string
}

func (p *AnthropicProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
    params := anthropic.MessageNewParams{
        Model:     anthropic.Model(p.model),
        Messages:  convertMessages(req.Messages),
        MaxTokens: int64(req.MaxTokens),
    }
    if len(req.Tools) > 0 {
        params.Tools = param.NewOpt(convertTools(req.Tools))
    }
    resp, err := p.client.Messages.New(ctx, params)
    return convertResponse(resp), err
}
```

Key: `param.NewOpt(v)` wraps optional fields — Anthropic SDK requires explicit opt-in.

---

## Ollama Provider

```go
type OllamaProvider struct {
    baseURL string
    model   string
}

func (p *OllamaProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
    // POST /api/chat with Ollama JSON format
    // Supports tool use via function_call format
}
```

### OllamaEmbedder

```go
type OllamaEmbedder struct {
    baseURL string
    model   string  // mxbai-embed-large → 1024-dim (Voyage AI v3 compatible)
}

func (e *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
    // POST /api/embed → returns vector(1024)
}
```

---

## Health Endpoint

```
GET /api/health/llm
```

Returns SmartRouter pool status:

```json
{
  "providers": [
    { "name": "ollama-heavy", "model": "qwen3:14b", "status": "ok", "circuit": "closed" },
    { "name": "ollama-fast",  "model": "qwen3:4b",  "status": "ok", "circuit": "closed" },
    { "name": "anthropic",    "model": "claude-sonnet-4-6", "status": "ok", "circuit": "closed" }
  ],
  "healthy": true
}
```

The Topnav LLM status dot polls this endpoint every 60 seconds.

---

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `LLM_PROVIDER` | `ollama` | `anthropic`, `ollama`, or `routing` |
| `ANTHROPIC_API_KEY` | — | Required when provider includes Anthropic |
| `OLLAMA_BASE_URL` | `http://host.docker.internal:11434` | Local Ollama endpoint |
| `OLLAMA_HEAVY_MODEL` | `qwen3:14b` | Primary tool-use model |
| `OLLAMA_FAST_MODEL` | `qwen3:4b` | Lightweight fallback |
| `OLLAMA_EMBED_MODEL` | `mxbai-embed-large` | Embedding model (1024-dim) |
| `OLLAMA_CLOUD_URL` | — | Remote Ollama endpoint |
| `OLLAMA_CLOUD_MODEL` | — | e.g. `deepseek-v3.2:cloud` |
| `OLLAMA_CLOUD_API_KEY` | — | Bearer token from ollama.com |

### buildSmartRouter (main.go)

```go
func buildSmartRouter(cfg *config.Config) llm.Provider {
    var providers []llm.ProviderWithCapabilities

    if cfg.OllamaBaseURL != "" {
        providers = append(providers,
            llm.WithCapabilities(
                llm.NewOllamaProvider(cfg.OllamaBaseURL, cfg.OllamaHeavyModel),
                []string{"tool-use", "heavy"},
            ),
            llm.WithCapabilities(
                llm.NewOllamaProvider(cfg.OllamaBaseURL, cfg.OllamaFastModel),
                []string{"tool-use", "fast"},
            ),
        )
    }

    if cfg.AnthropicAPIKey != "" {
        providers = append(providers,
            llm.WithCapabilities(
                llm.NewAnthropicProvider(cfg.AnthropicAPIKey, "claude-sonnet-4-6"),
                []string{"tool-use", "heavy", "fast"},
            ),
        )
    }

    return llm.NewSmartRouter(providers)
}
```
