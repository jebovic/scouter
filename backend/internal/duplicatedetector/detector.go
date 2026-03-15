package duplicatedetector

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/jibei/scouter/internal/shopping"
)

// frenchStopWords is the set of common French words excluded from token comparison.
var frenchStopWords = map[string]struct{}{
	"le": {}, "la": {}, "les": {}, "de": {}, "du": {}, "un": {},
	"une": {}, "pour": {}, "avec": {}, "en": {}, "et": {}, "ou": {},
	"à": {}, "au": {}, "aux": {}, "sur": {}, "par": {}, "des": {},
}

// normalize returns a lowercase, punctuation-free version of s.
func normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	return strings.TrimSpace(b.String())
}

// tokenize splits a normalized name into meaningful words (stop words removed).
func tokenize(name string) map[string]struct{} {
	tokens := make(map[string]struct{})
	for _, word := range strings.Fields(normalize(name)) {
		if _, stop := frenchStopWords[word]; !stop && len(word) > 1 {
			tokens[word] = struct{}{}
		}
	}
	return tokens
}

// jaccard returns the Jaccard similarity coefficient between two token sets.
func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	intersection := 0
	for k := range a {
		if _, ok := b[k]; ok {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// recommend returns the itemID that should be kept from a group.
// Preference order: status "buy" first, then lowest price.
func recommend(items []DuplicateItem) string {
	if len(items) == 0 {
		return ""
	}
	best := items[0]
	for _, item := range items[1:] {
		if item.Status == "buy" && best.Status != "buy" {
			best = item
			continue
		}
		if best.Status == "buy" {
			continue
		}
		if item.Price < best.Price {
			best = item
		}
	}
	return best.ItemID
}

// buildReason returns a French explanation for a duplicate group.
func buildReason(sim float64, commonWords int) string {
	pct := int(math.Round(sim * 100))
	return fmt.Sprintf("Ces articles semblent identiques (%d%% de mots en commun)", pct)
}

// Detect finds groups of likely-duplicate shopping items based on Jaccard similarity.
// Items with similarity >= 0.5 are grouped together.
func Detect(items []shopping.Item) []DuplicateGroup {
	if len(items) < 2 {
		return nil
	}

	// Precompute token sets for each item.
	tokens := make([]map[string]struct{}, len(items))
	for i, it := range items {
		tokens[i] = tokenize(it.Name)
	}

	// Union-find to group similar items.
	parent := make([]int, len(items))
	for i := range parent {
		parent[i] = i
	}
	simMap := make(map[[2]int]float64)

	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(x, y int) {
		px, py := find(x), find(y)
		if px != py {
			parent[px] = py
		}
	}

	const threshold = 0.5
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			sim := jaccard(tokens[i], tokens[j])
			if sim >= threshold {
				simMap[[2]int{i, j}] = sim
				union(i, j)
			}
		}
	}

	// Collect groups (only groups with > 1 member).
	groupMap := make(map[int][]int) // root → member indices
	for i := range items {
		root := find(i)
		groupMap[root] = append(groupMap[root], i)
	}

	var groups []DuplicateGroup
	for _, members := range groupMap {
		if len(members) < 2 {
			continue
		}
		// Compute the average similarity across pairs in the group.
		pairCount := 0
		totalSim := 0.0
		for a := 0; a < len(members); a++ {
			for b := a + 1; b < len(members); b++ {
				key := [2]int{members[a], members[b]}
				if members[a] > members[b] {
					key = [2]int{members[b], members[a]}
				}
				if s, ok := simMap[key]; ok {
					totalSim += s
					pairCount++
				}
			}
		}
		avgSim := 0.0
		if pairCount > 0 {
			avgSim = totalSim / float64(pairCount)
		}

		// Build DuplicateItem list.
		dupeItems := make([]DuplicateItem, 0, len(members))
		for _, idx := range members {
			it := items[idx]
			dupeItems = append(dupeItems, DuplicateItem{
				ItemID:   it.ID.String(),
				ItemName: it.Name,
				Price:    it.Price,
				Status:   it.Status,
				Merchant: it.Merchant,
			})
		}

		commonWords := 0
		if pairCount > 0 {
			commonWords = int(math.Round(avgSim * float64(len(tokens[members[0]])+len(tokens[members[1]])) / 2))
		}

		groups = append(groups, DuplicateGroup{
			Items:       dupeItems,
			Similarity:  math.Round(avgSim*100) / 100,
			Reason:      buildReason(avgSim, commonWords),
			Recommended: recommend(dupeItems),
		})
	}

	return groups
}
