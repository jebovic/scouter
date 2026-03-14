package agentrun

import "strings"

// Truncate shortens s to at most max runes (not bytes), preserving valid UTF-8.
func Truncate(s string, max int) string {
	n := 0
	for i := range s {
		if n == max {
			return s[:i]
		}
		n++
	}
	return s
}

// SanitizeFeedback strips ASCII control characters (except newline and tab),
// collapses triple newlines to double, trims surrounding whitespace,
// and truncates to max runes. Use before interpolating user text into LLM prompts.
func SanitizeFeedback(s string, max int) string {
	var b strings.Builder
	for _, r := range s {
		if r < 0x20 && r != '\n' && r != '\t' {
			continue
		}
		b.WriteRune(r)
	}
	out := b.String()
	for strings.Contains(out, "\n\n\n") {
		out = strings.ReplaceAll(out, "\n\n\n", "\n\n")
	}
	return Truncate(strings.TrimSpace(out), max)
}
