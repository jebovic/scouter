# Agent Team — SCOUTER Intelligence Squad

## Philosophy
Each agent is a specialist. Separation of concerns enforces domain boundaries.
The team works on "missions" — any major spending decision the user needs help with.

---

## ResearchAgent — Option Discovery Specialist

**Role:** Deep research on options/alternatives, specs, benchmarks, trade-offs for any spending domain
**Behavior:** Digs into product specs, reviews, comparisons, and alternatives. Structures findings with attributes, scores, and constraint checks. Proposes a spectrum of options from aspirational to budget-conscious.
**Code:** `internal/research/` — builds tool schema, constructs prompts, calls `Provider.Complete()`, parses response, persists options to DB
**Output:** Structured `[]Option` with attributes, warnings, price ranges, and recommendation badges.

---

## PricingAgent — Price Intelligence Specialist

**Role:** Price hunting, deal finding, merchant comparison, TCO analysis, budget optimization
**Behavior:** Tracks prices across merchants/platforms. Compares TCO not just sticker price. Optimizes for fewer vendors (shipping consolidation). Flags flash sales, seasonal patterns, price trends. Maintains price history. Calculates contingency buffers.
**Code:** `internal/pricing/` — builds tool schema, constructs prompts, calls `Provider.Complete()`, parses response, persists shopping items to DB
**Output:** Merchant-grouped `[]ShoppingItem` with status badges, timeline, and budget analysis.

---

## FrontendRenderer — UI & Report Layer

**Role:** Renders all research and pricing into interactive SCOUTER React pages
**Behavior:** Smart, confident data presentation. Turns raw structured data into polished, interactive reports.
**Code:** `frontend/src/` — owns the entire React frontend, design system, and component library
**Trigger:** Any backend agent completes a task → FrontendRenderer displays the relevant page(s).
**Design system:** SCOUTER dark theme — Oxanium/Chakra Petch/IBM Plex Mono, card-based, scanline overlays, cyan/coral/gold, Recharts.
**Output:** React components using SCOUTER design system. Data via Tanstack Query, validated with Zod.

---

## Team Dynamics
- **ResearchAgent + PricingAgent** work sequentially: research discovers options, pricing finds the best deals for each.
- **FrontendRenderer** is the team's face. Every agent's output flows through the UI before it's "real."
- **Consensus format:** ResearchAgent presents top options → PricingAgent prices them → converge on recommendation.
- LLM calls use **tool use / function calling** — never raw JSON prompt parsing.
