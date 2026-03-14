package decision

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
)

// Score computes a ScoreResult for every option.
// It is pure (no I/O) and deterministic given the same inputs.
func Score(m mission.Mission, opts []option.Option) []ScoreResult {
	wp := normalizeWeights(m.WeightProfile)
	results := make([]ScoreResult, 0, len(opts))
	for _, o := range opts {
		results = append(results, scoreOne(m, o, wp))
	}
	return results
}

type weights struct{ price, qual, feat float64 }

func normalizeWeights(wp mission.WeightProfile) weights {
	p, q, f := wp.Price, wp.Quality, wp.Feature
	total := p + q + f
	if total <= 0 {
		// default: equal weight
		return weights{price: 1.0 / 3, qual: 1.0 / 3, feat: 1.0 / 3}
	}
	return weights{price: p / total, qual: q / total, feat: f / total}
}

func scoreOne(m mission.Mission, o option.Option, w weights) ScoreResult {
	r := ScoreResult{
		OptionID:   o.ID,
		OptionName: o.Name,
	}

	// Hard constraint check
	for _, c := range m.Constraints {
		if c.Type != "hard" {
			continue
		}
		for _, attr := range o.Attributes {
			if attr.Key != c.Key {
				continue
			}
			if attr.Pass != nil && !*attr.Pass {
				r.Eliminated = true
				r.EliminReason = fmt.Sprintf("hard constraint %q not met", c.Label)
				return r
			}
		}
	}

	r.PriceScore = priceScore(m.Budget, o)
	r.QualScore = qualityScore(o)
	r.FeatScore = featureScore(m, o)

	composite := w.price*r.PriceScore + w.qual*r.QualScore + w.feat*r.FeatScore
	r.Score = clamp(composite, 0, 100)
	return r
}

// priceScore returns 0-100 based on how well the option fits the budget.
// Sweet-spot is 70-80% of budget (score=100). Over budget → penalized.
func priceScore(budget float64, o option.Option) float64 {
	if budget <= 0 || o.PriceRange == nil || o.PriceRange.Best <= 0 {
		return 50 // unknown price → neutral
	}
	ratio := o.PriceRange.Best / budget
	switch {
	case ratio <= 0.8:
		// Good value — scale 85-100 based on how close to sweet-spot
		return clamp(85+(0.8-ratio)*75, 85, 100)
	case ratio <= 1.0:
		// Slightly under budget — 70-85
		return clamp(85-(ratio-0.8)/0.2*15, 70, 85)
	default:
		// Over budget — penalize
		return clamp(70-(ratio-1.0)*150, 0, 70)
	}
}

// qualityScore returns 0-100 from score-type attributes (e.g., display 8/10).
func qualityScore(o option.Option) float64 {
	var total, count float64
	for _, attr := range o.Attributes {
		if attr.Type != "score" {
			continue
		}
		max := 10.0
		if attr.Max != nil && *attr.Max > 0 {
			max = float64(*attr.Max)
		}
		v := toFloat(attr.Value)
		total += clamp(v/max*100, 0, 100)
		count++
	}
	if count == 0 {
		return 50 // no quality data → neutral
	}
	return total / count
}

// featureScore returns 0-100 based on fraction of soft-constraint attributes that pass.
func featureScore(m mission.Mission, o option.Option) float64 {
	// Count soft constraints that have matching attributes.
	var checked, passed float64
	for _, c := range m.Constraints {
		if c.Type != "soft" {
			continue
		}
		for _, attr := range o.Attributes {
			if attr.Key != c.Key {
				continue
			}
			checked++
			if attr.Pass == nil || *attr.Pass {
				passed++
			}
		}
	}
	if checked == 0 {
		// No soft constraints — use boolean attributes as a proxy
		return booleanPassRate(o)
	}
	return (passed / checked) * 100
}

func booleanPassRate(o option.Option) float64 {
	var total, trueCount float64
	for _, attr := range o.Attributes {
		if attr.Type != "boolean" {
			continue
		}
		total++
		if b, ok := attr.Value.(bool); ok && b {
			trueCount++
		}
	}
	if total == 0 {
		return 50
	}
	return (trueCount / total) * 100
}

func toFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case json.Number:
		f, err := x.Float64()
		if err != nil {
			return 0
		}
		return f
	case string:
		f, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return 0
		}
		return f
	}
	return 0
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
