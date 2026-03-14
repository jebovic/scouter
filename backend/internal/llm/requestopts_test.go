package llm

import (
	"context"
	"testing"
)

func TestGetRequestOpts_DefaultWhenAbsent(t *testing.T) {
	ctx := context.Background()
	opts := GetRequestOpts(ctx)
	if opts.Capabilities != CapText {
		t.Errorf("expected default capabilities CapText (%d), got %d", CapText, opts.Capabilities)
	}
	if opts.Label != "" {
		t.Errorf("expected empty label, got %q", opts.Label)
	}
}

func TestWithRequestOpts_RoundTrip(t *testing.T) {
	original := RequestOpts{
		Capabilities: CapToolUse | CapText,
		Label:        "research",
	}
	ctx := WithRequestOpts(context.Background(), original)
	got := GetRequestOpts(ctx)
	if got.Capabilities != original.Capabilities {
		t.Errorf("capabilities: want %d, got %d", original.Capabilities, got.Capabilities)
	}
	if got.Label != original.Label {
		t.Errorf("label: want %q, got %q", original.Label, got.Label)
	}
}

func TestWithRequestOpts_DoesNotModifyParentContext(t *testing.T) {
	parent := context.Background()
	child := WithRequestOpts(parent, RequestOpts{Capabilities: CapToolUse, Label: "pricing"})

	parentOpts := GetRequestOpts(parent)
	childOpts := GetRequestOpts(child)

	if parentOpts.Capabilities != CapText {
		t.Errorf("parent context must remain unaffected, got capabilities %d", parentOpts.Capabilities)
	}
	if childOpts.Label != "pricing" {
		t.Errorf("child label: want %q, got %q", "pricing", childOpts.Label)
	}
}

func TestCapability_Has(t *testing.T) {
	tests := []struct {
		name     string
		c        Capability
		required Capability
		want     bool
	}{
		{
			name:     "CapText has CapText",
			c:        CapText,
			required: CapText,
			want:     true,
		},
		{
			name:     "CapToolUse does not have CapText",
			c:        CapToolUse,
			required: CapText,
			want:     false,
		},
		{
			name:     "CapText|CapToolUse has CapToolUse",
			c:        CapText | CapToolUse,
			required: CapToolUse,
			want:     true,
		},
		{
			name:     "CapText|CapToolUse has CapText",
			c:        CapText | CapToolUse,
			required: CapText,
			want:     true,
		},
		{
			name:     "zero capability has CapText returns false",
			c:        0,
			required: CapText,
			want:     false,
		},
		{
			name:     "zero capability has zero returns true (vacuously)",
			c:        0,
			required: 0,
			want:     true,
		},
		{
			name:     "all caps have CapLongCtx",
			c:        CapText | CapToolUse | CapLongCtx,
			required: CapLongCtx,
			want:     true,
		},
		{
			name:     "CapText does not have CapLongCtx",
			c:        CapText,
			required: CapLongCtx,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Has(tt.required)
			if got != tt.want {
				t.Errorf("(%d).Has(%d) = %v, want %v", tt.c, tt.required, got, tt.want)
			}
		})
	}
}

func TestCapability_BitwiseIsolation(t *testing.T) {
	// Ensure each constant has a unique bit
	if CapText&CapToolUse != 0 {
		t.Error("CapText and CapToolUse share bits")
	}
	if CapText&CapLongCtx != 0 {
		t.Error("CapText and CapLongCtx share bits")
	}
	if CapToolUse&CapLongCtx != 0 {
		t.Error("CapToolUse and CapLongCtx share bits")
	}
}
