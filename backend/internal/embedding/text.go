// Package embedding handles vector embedding generation and persistence for options.
// It is responsible for building embed-ready text from option data, storing vectors,
// and running the async worker that keeps embeddings up to date.
package embedding

import (
	"fmt"
	"strings"

	"github.com/jibei/scouter/internal/option"
)

// BuildEmbedText constructs a short, dense text representation of an option for
// embedding. It concatenates the name, category, notes, and key attribute values
// so that semantic search surfaces options by content, not by IDs.
func BuildEmbedText(opt option.Option) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s. Category: %s.", opt.Name, opt.Category)
	if opt.Notes != "" {
		fmt.Fprintf(&sb, " %s", opt.Notes)
	}
	for _, a := range opt.Attributes {
		fmt.Fprintf(&sb, " %s: %v.", a.Label, a.Value)
	}
	return strings.TrimSpace(sb.String())
}

// FormatVector serialises a float64 slice into the Postgres vector literal syntax
// "[a,b,c]" so it can be cast with ::vector in SQL without the pgvector-go library.
func FormatVector(vec []float64) string {
	if len(vec) == 0 {
		return "[]"
	}
	parts := make([]string, len(vec))
	for i, v := range vec {
		parts[i] = fmt.Sprintf("%g", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
