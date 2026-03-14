package llm

import "context"

// Capability is a bitmask describing what an LLM model supports.
type Capability uint8

const (
	// CapText marks models that produce plain text output.
	CapText Capability = 1 << iota
	// CapToolUse marks models that support function/tool calling.
	CapToolUse
	// CapLongCtx marks models with an extended context window (>32k tokens).
	CapLongCtx
)

// Has reports whether c includes all bits in required.
func (c Capability) Has(required Capability) bool {
	return c&required == required
}

// RequestOpts carries per-request routing hints attached to a context.
type RequestOpts struct {
	// Capabilities lists the minimum capabilities required for this request.
	Capabilities Capability
	// Label is a short human-readable tag used for logging/metrics.
	Label string
}

type ctxKey struct{}

// WithRequestOpts returns a new context carrying opts.
// It does not mutate the parent context.
func WithRequestOpts(ctx context.Context, opts RequestOpts) context.Context {
	return context.WithValue(ctx, ctxKey{}, opts)
}

// GetRequestOpts retrieves RequestOpts from ctx.
// Returns defaults (CapText, empty label) when none are attached.
func GetRequestOpts(ctx context.Context) RequestOpts {
	if v, ok := ctx.Value(ctxKey{}).(RequestOpts); ok {
		return v
	}
	return RequestOpts{Capabilities: CapText}
}

// HasRequestOpts reports whether a RequestOpts was explicitly attached to ctx.
func HasRequestOpts(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKey{}).(RequestOpts)
	return ok
}
