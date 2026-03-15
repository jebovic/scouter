package reviewsummary

import (
	"fmt"
	"hash/fnv"
	"time"
)

// prosPool and consPool are candidate phrases indexed by (seed % len).
var prosPool = []string{
	"Excellent build quality",
	"Great value for the price",
	"Easy to set up and use",
	"Long-lasting battery life",
	"Responsive customer support",
	"Compact and lightweight design",
	"Accurate performance metrics",
	"Wide compatibility with accessories",
	"Quiet operation",
	"Durable materials",
	"Intuitive user interface",
	"Fast charging support",
	"Solid warranty coverage",
	"Energy efficient",
	"Good software ecosystem",
}

var consPool = []string{
	"Higher price than competitors",
	"Limited color options",
	"Manual could be clearer",
	"Occasional connectivity issues",
	"No included case or bag",
	"App requires account registration",
	"Heavy for extended use",
	"Proprietary charging port",
	"Short cable included",
	"Limited storage options",
}

var verdictPool = []string{
	"A reliable choice for most buyers — the pros outweigh the minor drawbacks.",
	"Highly recommended for daily use; best-in-class at this price point.",
	"Solid performer with a few rough edges, but overall worth the investment.",
	"Good mid-range option; power users may want to look at premium alternatives.",
	"Exceptional quality and value — a crowd-pleaser across all user segments.",
}

var recommendedForPool = []string{
	"Budget-conscious shoppers",
	"First-time buyers",
	"Power users who need reliability",
	"Home office setups",
	"Frequent travellers",
	"Students and young professionals",
	"Households with multiple users",
}

var avoidIfPool = []string{
	"You need advanced pro features",
	"Brand loyalty to a competing ecosystem",
	"You require a very slim form factor",
	"You expect white-glove enterprise support",
	"You need extensive third-party integrations",
}

// generate builds a deterministic ReviewSummary from a 32-bit seed.
func generate(itemID, itemName string) *ReviewSummary {
	h := fnv.New32a()
	h.Write([]byte(itemID))
	seed := h.Sum32()

	// Overall score: 6.5–9.5 with one decimal place.
	scoreStep := int(seed%31) // 0..30
	overallScore := 6.5 + float64(scoreStep)*0.1

	// Total reviews: 120–12 000.
	totalReviews := 120 + int(seed%11881)

	// Pick 3–5 pros.
	nPros := 3 + int(seed%3)
	pros := make([]string, 0, nPros)
	for i := 0; i < nPros; i++ {
		idx := (int(seed) + i*7) % len(prosPool)
		pros = append(pros, prosPool[idx])
	}

	// Pick 2–3 cons.
	nCons := 2 + int(seed%2)
	cons := make([]string, 0, nCons)
	for i := 0; i < nCons; i++ {
		idx := (int(seed) + i*13) % len(consPool)
		cons = append(cons, consPool[idx])
	}

	verdict := verdictPool[int(seed)%len(verdictPool)]

	// Pick 2–3 recommendedFor entries.
	nRec := 2 + int(seed%2)
	recommendedFor := make([]string, 0, nRec)
	for i := 0; i < nRec; i++ {
		idx := (int(seed) + i*5) % len(recommendedForPool)
		recommendedFor = append(recommendedFor, recommendedForPool[idx])
	}

	// Pick 1–2 avoidIf entries.
	nAvoid := 1 + int(seed%2)
	avoidIf := make([]string, 0, nAvoid)
	for i := 0; i < nAvoid; i++ {
		idx := (int(seed) + i*11) % len(avoidIfPool)
		avoidIf = append(avoidIf, avoidIfPool[idx])
	}

	displayName := itemName
	if displayName == "" {
		displayName = fmt.Sprintf("Item %s", itemID)
	}

	return &ReviewSummary{
		ItemID:         itemID,
		ItemName:       displayName,
		OverallScore:   overallScore,
		TotalReviews:   totalReviews,
		Pros:           pros,
		Cons:           cons,
		Verdict:        verdict,
		RecommendedFor: recommendedFor,
		AvoidIf:        avoidIf,
		LastUpdated:    time.Now().UTC(),
	}
}
