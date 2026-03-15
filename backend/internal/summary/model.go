// Package summary implements the AI Mission Summary Card (Phase 72).
// It generates a concise French-language executive brief via the LLM using
// tool-use for structured output. The result is cached in-memory (sync.RWMutex
// + 1h TTL) keyed by mission slug — no DB migration needed.
package summary

// MissionSummary is the AI-generated executive brief for a mission.
type MissionSummary struct {
	MissionSlug string   `json:"missionSlug"`
	Bullets     []string `json:"bullets"`  // 3-5 French bullet points with emojis
	Verdict     string   `json:"verdict"`  // "go" | "wait" | "research_more"
	CachedAt    int64    `json:"cachedAt"` // Unix timestamp (seconds)
}
