package comparison

import (
	"math"
	"sort"
)

// priceInverted lists criteria for which lower values are better.
// A score normalization for these criteria is inverted: 100 - normalizedRaw.
var priceInverted = map[string]bool{
	"price": true,
	"cost":  true,
}

// normalizeValue maps val within [min, max] to a 0–100 score.
// For price-like criteria the score is inverted (lower is better).
// When min == max, 50 is returned to avoid division by zero.
func normalizeValue(criteria string, val, min, max float64) float64 {
	if max == min {
		return 50
	}
	normalized := (val - min) / (max - min) * 100
	if priceInverted[criteria] {
		normalized = 100 - normalized
	}
	return math.Round(normalized*100) / 100
}

// computeScores calculates a weighted matrix score for each raw option.
// The algorithm:
//  1. For each weight criterion, collect all numeric values across options to
//     determine the global [min, max] range for that criterion.
//  2. Normalize each option's value to [0, 100], inverting price-like criteria.
//  3. Score = Σ(weight_i * normalized_i) / Σ(weight_i).
//     Missing / non-numeric attributes contribute 0.
//  4. Options are sorted descending by score; rank 1 = best.
func computeScores(opts []rawOption, weights []Weight) []OptionScore {
	if len(opts) == 0 {
		return []OptionScore{}
	}

	// Total weight sum.
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w.Weight
	}

	// If no weights defined, every option scores 0.
	if totalWeight == 0 {
		scores := make([]OptionScore, len(opts))
		for i, o := range opts {
			scores[i] = OptionScore{OptionID: o.id, OptionName: o.name, Score: 0, Rank: i + 1}
		}
		return scores
	}

	// Pre-compute min/max for each criterion across all options.
	type rangeEntry struct{ min, max float64 }
	ranges := map[string]*rangeEntry{}
	for _, w := range weights {
		ranges[w.Criteria] = &rangeEntry{min: math.MaxFloat64, max: -math.MaxFloat64}
	}
	for _, o := range opts {
		for _, w := range weights {
			v, ok := numericAttr(o.attrs, w.Criteria)
			if !ok {
				continue
			}
			r := ranges[w.Criteria]
			if v < r.min {
				r.min = v
			}
			if v > r.max {
				r.max = v
			}
		}
	}

	// Compute per-option scores.
	scores := make([]OptionScore, len(opts))
	for i, o := range opts {
		weighted := 0.0
		for _, w := range weights {
			if w.Weight == 0 {
				continue
			}
			v, ok := numericAttr(o.attrs, w.Criteria)
			if !ok {
				continue
			}
			r := ranges[w.Criteria]
			norm := normalizeValue(w.Criteria, v, r.min, r.max)
			weighted += float64(w.Weight) * norm
		}
		score := weighted / float64(totalWeight)
		scores[i] = OptionScore{
			OptionID:   o.id,
			OptionName: o.name,
			Score:      math.Round(score*100) / 100,
		}
	}

	// Sort descending by score; stable to give deterministic rank on tie.
	sort.SliceStable(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})
	for i := range scores {
		scores[i].Rank = i + 1
	}
	return scores
}

// numericAttr extracts a float64 from attrs[key] if the value is numeric.
func numericAttr(attrs map[string]interface{}, key string) (float64, bool) {
	v, ok := attrs[key]
	if !ok {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return t, true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case float32:
		return float64(t), true
	}
	return 0, false
}
