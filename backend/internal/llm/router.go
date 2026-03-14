package llm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// circuitBreaker is a simple 3-failure → 5-minute open circuit breaker.
type circuitBreaker struct {
	mu          sync.Mutex
	failures    int
	openUntil   time.Time
	maxFailures int
	openDur     time.Duration
}

func newCircuitBreaker() *circuitBreaker {
	return &circuitBreaker{maxFailures: 3, openDur: 5 * time.Minute}
}

// allow returns true if the circuit is closed (primary may be tried).
func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return time.Now().After(cb.openUntil)
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.openUntil = time.Time{}
}

func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.failures >= cb.maxFailures {
		cb.openUntil = time.Now().Add(cb.openDur)
	}
}

// state returns the human-readable circuit state: "open", "half_open", or "closed".
// "half_open" means the open window has expired and the next request may probe.
func (cb *circuitBreaker) state() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.openUntil.IsZero() {
		return "closed"
	}
	if time.Now().After(cb.openUntil) {
		return "half_open"
	}
	return "open"
}

// isInfraError returns true only for infrastructure-level failures that justify
// falling back to the secondary provider. Model-quality failures (4xx, bad JSON,
// tool parse errors) are NOT infrastructure errors and should be returned to the
// caller without burning secondary provider tokens.
func isInfraError(err error) bool {
	// HTTP 5xx from upstream (e.g. Ollama overloaded)
	var se *StatusError
	if errors.As(err, &se) {
		return se.Code >= 500
	}
	// TCP-level connection failure (connection refused, DNS, etc.)
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	// HTTP client timeout
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return false
}

// RoutingProvider routes requests to a primary provider (Ollama) with strict
// automatic fallback to a secondary provider (Anthropic) only on infrastructure
// failures: connection errors, client timeouts, or HTTP 5xx responses.
//
// Rate limits: primary 60 req/min, secondary 5 req/min.
// Circuit breaker: 3 consecutive infra failures → primary bypassed for 5 min.
type RoutingProvider struct {
	primary   Provider
	secondary Provider
	cb        *circuitBreaker
	primaryRL *rate.Limiter
	secondRL  *rate.Limiter
}

// NewRoutingProvider creates a RoutingProvider.
func NewRoutingProvider(primary, secondary Provider) *RoutingProvider {
	return &RoutingProvider{
		primary:   primary,
		secondary: secondary,
		cb:        newCircuitBreaker(),
		primaryRL: rate.NewLimiter(rate.Every(time.Minute/60), 10),
		secondRL:  rate.NewLimiter(rate.Every(time.Minute/5), 2),
	}
}

// Complete routes to primary when the circuit is closed. Falls back to secondary
// only on infrastructure failures. Returns the primary error directly for all
// non-infrastructure failures (model quality, auth, bad JSON, 4xx).
func (p *RoutingProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	if p.cb.allow() {
		if err := p.primaryRL.Wait(ctx); err != nil {
			return CompletionResponse{}, fmt.Errorf("primary rate limit: %w", err)
		}
		resp, err := p.primary.Complete(ctx, req)
		if err == nil {
			p.cb.recordSuccess()
			return resp, nil
		}
		if !isInfraError(err) {
			// Model-quality failure — return directly, do not burn secondary tokens.
			return CompletionResponse{}, err
		}
		p.cb.recordFailure()
		// Infra failure: fall through to secondary.
	}

	if err := p.secondRL.Wait(ctx); err != nil {
		return CompletionResponse{}, fmt.Errorf("secondary rate limit: %w", err)
	}
	resp, err := p.secondary.Complete(ctx, req)
	if err != nil {
		return CompletionResponse{}, err
	}
	resp.WasFallback = true
	return resp, nil
}
