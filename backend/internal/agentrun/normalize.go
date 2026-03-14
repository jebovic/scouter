package agentrun

import (
	"strings"
	"unicode"
)

// normalizeName lowercases, trims whitespace, and strips trademark/copyright symbols
// from a product name to improve matching across LLM name variations.
func normalizeName(name string) string {
	// Strip trademark symbols and parenthetical suffixes like "(2024)", "(128GB)"
	name = strings.Map(func(r rune) rune {
		switch r {
		case '™', '®', '©':
			return -1
		}
		return r
	}, name)

	// Remove parenthetical suffixes
	if i := strings.IndexRune(name, '('); i > 0 {
		name = name[:i]
	}

	// Normalize quotes and dashes to ASCII
	name = strings.ReplaceAll(name, "\u2019", "'")
	name = strings.ReplaceAll(name, "\u201c", "\"")
	name = strings.ReplaceAll(name, "\u201d", "\"")

	// Collapse whitespace
	var b strings.Builder
	prevSpace := true
	for _, r := range strings.ToLower(name) {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteRune(' ')
			}
			prevSpace = true
		} else {
			b.WriteRune(r)
			prevSpace = false
		}
	}

	return strings.TrimSpace(b.String())
}

// ComputeDiff computes the diff between a previous and current result snapshot.
// Matching is done by normalized name. Items in prev but not in curr are "removed";
// items in curr but not in prev are "added"; items present in both with the same
// name are not returned (unchanged). "changed" is not currently used at this tier
// (reserved for future attribute-level diff).
func ComputeDiff(prev, curr []SnapshotItem) Diff {
	d := Diff{
		Added:   []DiffEntry{},
		Removed: []DiffEntry{},
		Changed: []DiffEntry{},
	}

	// Build a normalized lookup of prev items.
	prevByNorm := make(map[string]SnapshotItem, len(prev))
	for _, item := range prev {
		prevByNorm[normalizeName(item.Name)] = item
	}

	// Build a normalized lookup of curr items.
	currByNorm := make(map[string]SnapshotItem, len(curr))
	for _, item := range curr {
		currByNorm[normalizeName(item.Name)] = item
	}

	// Items in curr not in prev → added.
	for norm, item := range currByNorm {
		if _, exists := prevByNorm[norm]; !exists {
			d.Added = append(d.Added, DiffEntry{ID: item.ID, Name: item.Name, Status: "new"})
		}
	}

	// Items in prev not in curr → removed.
	for norm, item := range prevByNorm {
		if _, exists := currByNorm[norm]; !exists {
			d.Removed = append(d.Removed, DiffEntry{ID: item.ID, Name: item.Name, Status: "removed"})
		}
	}

	return d
}
