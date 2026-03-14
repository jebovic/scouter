package agentrun

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
