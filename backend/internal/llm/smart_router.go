package llm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/time/rate"
)

// LLMRecorder is a subset of metrics.LLMRecorder used by SmartRouter,
// defined here to avoid an import cycle between llm and metrics.
type LLMRecorder interface {
	RecordLLMCall(model, label, result string, duration time.Duration)
	RecordLLMTokens(model, direction string, count int)
	RecordLLMFallback(model string)
}

// noopLLMRecorder is used when no recorder is provided.
type noopLLMRecorder struct{}

func (noopLLMRecorder) RecordLLMCall(_, _, _ string, _ time.Duration) {}
func (noopLLMRecorder) RecordLLMTokens(_, _ string, _ int)            {}
func (noopLLMRecorder) RecordLLMFallback(_ string)                    {}

// ErrAllModelsUnavailable is returned when every model in the pool either has an
// open circuit breaker or failed with an infrastructure error.
var ErrAllModelsUnavailable = errors.New("llm: all models unavailable")

// SmartRouter routes LLM requests across a ModelPool using capability-based
// selection, per-model circuit breakers, per-model timeouts, and optional
// per-model rate limits.
//
// Models are tried in Priority order (lowest value first). Cascade continues
// only on infrastructure errors (isInfraError). Any non-infra error (4xx, auth,
// bad JSON) is returned immediately without trying additional models.
type SmartRouter struct {
	pool     ModelPool
	breakers map[string]*circuitBreaker
	limiters map[string]*rate.Limiter
	log      *slog.Logger
	recorder LLMRecorder
}

// SmartRouterOption is a functional option for NewSmartRouter.
type SmartRouterOption func(*SmartRouter)

// WithModelRateLimit adds a requests-per-minute rate limiter for a named model.
// Useful for cloud models with strict RPM quotas.
func WithModelRateLimit(modelName string, rpm int) SmartRouterOption {
	return func(r *SmartRouter) {
		if rpm > 0 {
			burst := rpm/10 + 1 // ~10% of rpm; +1 ensures ≥1 token for low-rpm models
			r.limiters[modelName] = rate.NewLimiter(rate.Every(time.Minute/time.Duration(rpm)), burst)
		}
	}
}

// WithRecorder sets the LLMRecorder for the SmartRouter to instrument calls.
func WithRecorder(rec LLMRecorder) SmartRouterOption {
	return func(r *SmartRouter) {
		if rec != nil {
			r.recorder = rec
		}
	}
}

// NewSmartRouter creates a SmartRouter with one circuit breaker per pool entry.
func NewSmartRouter(pool ModelPool, log *slog.Logger, opts ...SmartRouterOption) *SmartRouter {
	breakers := make(map[string]*circuitBreaker, len(pool))
	for _, e := range pool {
		breakers[e.Name] = newCircuitBreaker()
	}
	sr := &SmartRouter{
		pool:     pool,
		breakers: breakers,
		limiters: make(map[string]*rate.Limiter),
		log:      log,
		recorder: noopLLMRecorder{},
	}
	for _, opt := range opts {
		opt(sr)
	}
	return sr
}

// Pool exposes the model pool for health checking.
func (r *SmartRouter) Pool() ModelPool {
	return r.pool
}

// Breakers exposes the per-model circuit breaker map for health checking.
func (r *SmartRouter) Breakers() map[string]*circuitBreaker {
	return r.breakers
}

// Complete selects models by capability, then tries each in priority order,
// cascading to the next only on infrastructure failures.
func (r *SmartRouter) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	if !HasRequestOpts(ctx) {
		r.log.WarnContext(ctx, "SmartRouter: no RequestOpts in context, defaulting to CapText")
	}
	opts := GetRequestOpts(ctx)

	candidates := r.pool.ForCapabilities(opts.Capabilities)
	if len(candidates) == 0 {
		r.log.WarnContext(ctx, "SmartRouter: no models match capabilities",
			"capabilities", opts.Capabilities,
			"label", opts.Label)
		return CompletionResponse{}, ErrAllModelsUnavailable
	}

	attempts := 0
	for _, entry := range candidates {
		cb := r.breakers[entry.Name]
		if cb != nil && !cb.allow() {
			r.log.DebugContext(ctx, "SmartRouter: circuit open, skipping model",
				"model", entry.Name,
				"label", opts.Label)
			continue
		}

		// Apply per-model rate limit if configured.
		if lim, ok := r.limiters[entry.Name]; ok {
			if err := lim.Wait(ctx); err != nil {
				return CompletionResponse{}, fmt.Errorf("rate limit wait for %s: %w", entry.Name, err)
			}
		}

		// Apply per-model timeout.
		callCtx := ctx
		var cancel context.CancelFunc
		if entry.Timeout > 0 {
			callCtx, cancel = context.WithTimeout(ctx, entry.Timeout)
		}

		// Inject system prompt prefix (e.g. "/no_think" for Qwen3).
		callReq := req
		if entry.SystemPromptPrefix != "" {
			callReq = injectSystemPrefix(req, entry.SystemPromptPrefix)
		}

		attempts++
		start := time.Now()
		resp, err := entry.Provider.Complete(callCtx, callReq)
		latency := time.Since(start)
		if cancel != nil {
			cancel()
		}

		if err == nil {
			if cb != nil {
				cb.recordSuccess()
			}
			resp.ModelName = entry.Name
			resp.Attempts = attempts
			resp.Degraded = attempts > 1
			if attempts > 1 {
				resp.WasFallback = true
			}
			r.log.InfoContext(ctx, "SmartRouter: request completed",
				"model", entry.Name,
				"latency_ms", latency.Milliseconds(),
				"attempts", attempts,
				"degraded", resp.Degraded,
				"label", opts.Label)
			r.recorder.RecordLLMCall(entry.Name, opts.Label, "success", latency)
			if resp.Usage.InputTokens > 0 {
				r.recorder.RecordLLMTokens(entry.Name, "input", resp.Usage.InputTokens)
			}
			if resp.Usage.OutputTokens > 0 {
				r.recorder.RecordLLMTokens(entry.Name, "output", resp.Usage.OutputTokens)
			}
			return resp, nil
		}

		// Non-infra error: return immediately without cascading.
		if !isInfraError(err) {
			r.log.WarnContext(ctx, "SmartRouter: non-infra error, not cascading",
				"model", entry.Name,
				"err", err,
				"label", opts.Label)
			r.recorder.RecordLLMCall(entry.Name, opts.Label, "error", latency)
			return CompletionResponse{}, err
		}

		// Infra error: record failure, try next model.
		if cb != nil {
			cb.recordFailure()
		}
		r.log.WarnContext(ctx, "SmartRouter: infra error, cascading to next model",
			"model", entry.Name,
			"latency_ms", latency.Milliseconds(),
			"err", err,
			"label", opts.Label)
		r.recorder.RecordLLMCall(entry.Name, opts.Label, "infra_error", latency)
		r.recorder.RecordLLMFallback(entry.Name)
	}

	if attempts == 0 {
		r.log.WarnContext(ctx, "SmartRouter: all circuits open",
			"candidates", len(candidates),
			"label", opts.Label)
	} else {
		r.log.ErrorContext(ctx, "SmartRouter: all models exhausted after infra errors",
			"attempts", attempts,
			"label", opts.Label)
	}
	return CompletionResponse{}, ErrAllModelsUnavailable
}

// injectSystemPrefix prepends prefix as a system message, or prepends it to an
// existing system message's content if one is already first in the message list.
func injectSystemPrefix(req CompletionRequest, prefix string) CompletionRequest {
	msgs := make([]Message, 0, len(req.Messages)+1)
	if len(req.Messages) > 0 && req.Messages[0].Role == "system" {
		msgs = append(msgs, Message{
			Role:    "system",
			Content: prefix + "\n" + req.Messages[0].Content,
		})
		msgs = append(msgs, req.Messages[1:]...)
	} else {
		msgs = append(msgs, Message{Role: "system", Content: prefix})
		msgs = append(msgs, req.Messages...)
	}
	tools := make([]Tool, len(req.Tools))
	copy(tools, req.Tools)
	return CompletionRequest{
		Messages:  msgs,
		Tools:     tools,
		MaxTokens: req.MaxTokens,
	}
}
