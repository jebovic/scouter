# SCOUTER Universal — Bootstrap Prompt

> Feed this file to Claude Code in a fresh project directory to bootstrap the entire app.
> This is your initial briefing. Claude will create the project structure, build the shared design system, wire the reports, and populate sample data.

---

## Z Fighters Design Session

**Location:** Capsule Corp Lab, War Room
**Attendees:** Goku (Research Lead), Vegeta (Price Intelligence), Bulma (Systems Architect)
**Objective:** Design a general-purpose big-spending assistant that works for any major purchase category

---

### GOKU — Research Vision

Alright, here's the deal. The home server SCOUTER system was incredible — we had deep research on every CPU, every case, every motherboard. Benchmark comparisons, spec matrices, use-case fitness scoring. But it was locked to hardware.

What if I need to research holiday destinations the same way? I want to compare Crete vs Sardinia vs the Algarve with the same depth I compared the Ryzen 7700X vs the 7900. Flight options, hotel ratings, activity quality, weather data, kid-friendliness — all structured, all scorable, all comparable.

**The research engine needs to be domain-agnostic.** Instead of CPUs and cases, we need a universal concept: **Options**. Each option has attributes, and attributes have types — numeric (for benchmarking), categorical (for filtering), price (for budgeting), boolean (for pass/fail constraints). The Findings report becomes an **Options Explorer** — whatever you're researching, you can lay out alternatives with specs, score them against your criteria, and export winners to the shopping list.

I also want the concept of **Constraints** — hard requirements that filter out options immediately. For the home server it was "must fit in the living room" and "CMR only." For a holiday it might be "direct flight from Paris" and "max 4 hours." For a TV it's "OLED only" and "65 inches minimum." Constraints are the first thing you define in any mission, and they shape everything.

And missions. We need **concurrent missions**. I might be researching a TV at the same time as planning summer holidays. Each mission has its own budget, its own options, its own shopping list. The HQ dashboard shows all active missions side by side.

---

### VEGETA — Price Intelligence Vision

Goku's ideas are fine, but let me cut through the fluff. Here's what actually matters:

**Budget is king.** Every mission starts with a budget cap. Every option has a price. Every decision gets tracked against that cap in real time. The budget tracker from the Shopping report was the most useful thing we built — editable prices, live recalculation, percentage bars with color-coded status. That stays, but it needs to work for ANY currency and ANY budget size.

**TCO, not sticker price.** For the server we tracked shipping, flash sales, deferred purchases, and future upgrades separately. That's universal. A holiday has flights + hotel + activities + food + transport. A renovation has materials + labor + permits + unexpected costs. The shopping list needs **cost categories** that roll up into a total, with a contingency buffer built in.

**Merchant consolidation stays.** Whether it's Amazon vs LDLC for hardware, or Booking.com vs Airbnb for lodging, or Leroy Merlin vs Castorama for renovation materials — grouping by vendor and optimizing for fewer orders/bookings saves money. The merchant-grouped view was brilliant. Keep it.

**Deal status badges.** Buy, flash-sale, preorder, defer, watch, crisis, rejected — these map to any domain. A flight deal that expires tomorrow is a flash-sale. A TV you're monitoring for Black Friday is a watch. Materials with a 6-week lead time are a preorder. The badge system is already universal.

**Timeline.** Every big purchase has time pressure. Flash sales expire, restocks happen, seasonal prices shift, booking windows close. The timeline with urgency indicators stays. But it needs to be per-mission, not global.

One thing I insist on: **price history awareness.** We didn't have this in v1, and it burned us with the DDR5 crisis. Every item should optionally track price snapshots — when did it cost what. Even if it's just manual entry, having "I saw this at X price on Y date" is invaluable for knowing whether to buy now or wait.

---

### BULMA — Architecture Vision

Both of you are thinking about data. Let me think about the system.

**Architecture stays static.** No backend, no build tools, no frameworks. Static HTML + CSS + JS, Chart.js via CDN, localStorage for state. This is non-negotiable. It works offline, it works from a `file://` URL, it deploys anywhere. The moment you add a server, you add ops, and this is a personal tool, not a SaaS product.

**Multi-mission requires a new data model.** In v1 we had one implicit "mission" — the home server build. localStorage keys were flat: `scouter-day-one-total`, `scouter-findings-export`. For multi-mission, every key gets namespaced: `scouter-{missionId}-budget-total`, `scouter-{missionId}-options-export`, etc. The missionId is a slug generated from the mission name.

**The SCOUTER design system is fully portable.** Dark theme, scanline overlay, grid background, Oxanium/Chakra Petch/IBM Plex Mono fonts, card-based layout, cyan/coral/gold/green/purple/orange color tokens. All of this stays. I'll extract them into `shared/theme.css` so every report imports one file instead of duplicating 50 lines of color definitions.

**The reports are mission-agnostic shells.** The same `options/index.html` renders TV options or flight options or CPU options — it reads from a mission-scoped data object in localStorage and adapts its columns, charts, and filters accordingly.

---

## Convergence: The Specification

After debate, the three agents converge on the following specification.

---

## Project Identity

- **Name:** SCOUTER Universal
- **Tagline:** Your personal big-spending intelligence system.
- **Acronym:** Strategic Comparison & Optimization Unit for Total Expenditure Research
- **Version:** 2.0

## Core Concepts

| Concept | Description |
|---------|-------------|
| **Mission** | A single spending project (e.g., "Summer Holiday 2026", "Living Room TV") with its own budget, options, shopping list, and timeline |
| **Option** | A researched alternative within a mission (e.g., a destination, a TV model, a contractor quote) with structured attributes |
| **Constraint** | A hard or soft requirement that filters options (e.g., "OLED only", "direct flight", "max 2000 EUR") |
| **Shopping Item** | A concrete thing to buy/book/commit to, with price, merchant, status badge, and deal metadata |
| **Merchant** | A vendor/retailer/provider grouped for consolidation |
| **Timeline** | Dated milestones and deadlines with urgency indicators |
| **Price History** | Manual price snapshots for tracking trends over time |

## Tech Stack

- **Frontend:** React 19 + TypeScript (Vite), Tanstack Query for server state, React Router v7
- **Charts:** Recharts (React-native, no CDN dependency)
- **Design:** SCOUTER dark theme with sci-fi HUD aesthetic — Oxanium/Chakra Petch/IBM Plex Mono, CSS custom properties, scanline overlays, glow effects
- **i18n:** EN + FR via i18next
- **Backend:** Go 1.23+ with chi router, pgx/v5 for Postgres, structured JSON logging
- **Database:** PostgreSQL 16 + pgvector extension — relational tables for missions/options/shopping, jsonb for flexible attributes, vector columns for future semantic search/RAG
- **LLM:** Anthropic SDK (default, claude-sonnet-4-6) behind a Go interface — OllamaProvider ready for phase 2 (OpenAI-compatible API)
- **Deployment:** Docker Compose — `postgres`, `backend`, `frontend` services. `docker compose up` and done.

## File Structure

```
scouter-universal/
  CLAUDE.md                      -- Project instructions for Claude Code
  AGENTS.md                      -- Agent team roster and personas
  docker-compose.yml             -- postgres + backend + frontend
  .env.example                   -- required env vars (API keys, DB URL)

  backend/                       -- Go service
    cmd/server/main.go           -- entrypoint, env validation, router setup
    internal/
      config/                    -- env-based config struct
      db/                        -- pgx pool, migrations runner
      mission/                   -- handler, service, repository, model
      option/                    -- handler, service, repository, model
      shopping/                  -- handler, service, repository, model
      llm/                       -- LLMProvider interface + AnthropicProvider + OllamaProvider
      research/                  -- Goku agent logic (calls LLM, returns structured options)
      pricing/                   -- Vegeta agent logic (calls LLM, returns price intel)
    migrations/                  -- SQL migration files (001_init.sql, etc.)
    Dockerfile

  frontend/                      -- React + Vite + TS
    src/
      api/                       -- typed API client (fetch wrappers per resource)
      components/
        scouter/                 -- SCOUTER design system components (Card, Badge, BudgetBar, etc.)
        mission/                 -- MissionCard, MissionForm, ConstraintEditor
        options/                 -- OptionCard, AttributeRenderer, ComparisonTable
        shopping/                -- ShoppingList, MerchantGroup, PriceHistory
      pages/
        HQDashboard.tsx          -- mission selector + global budget overview
        MissionOverview.tsx      -- per-mission dashboard
        OptionsExplorer.tsx      -- research + comparison
        ShoppingTracker.tsx      -- budget + merchant-grouped list
      hooks/                     -- useMission, useOptions, useShopping, useResearch
      styles/
        theme.css                -- SCOUTER CSS custom properties (same tokens as v1)
      i18n/                      -- EN + FR translation files
      main.tsx
    Dockerfile
```

## Report System

### 1. HQ Dashboard (`reports/index.html`)
- Mission selector: create, switch between, archive missions
- Global budget overview across all active missions
- Per-mission status cards with phase indicator and spend percentage
- Quick-access tiles to each mission's reports
- Agent roster (cosmetic, Z Fighters branding)

### 2. Mission Overview (`reports/mission/overview/`)
- Per-mission dashboard: budget display, constraint list, phase timeline
- Summary stats: options explored, items in cart, budget utilization
- Recent activity log
- Mission settings (name, budget, currency, category)

### 3. Options Explorer (`reports/mission/options/`)
- Combines the old Hardware Report + Findings Report into one domain-agnostic view
- Option cards with flexible attribute rendering (adapts to any domain)
- Comparison mode: side-by-side attribute table for 2-3 options
- Constraint checker: green/red pass/fail badges per option
- Score visualization: radar charts for multi-attribute scoring
- Price range bars (min/max/best)
- Export to Shopping button per option
- Filter by category, badge, price range
- Collapsible sections with reveal animations

### 4. Shopping Tracker (`reports/mission/shopping/`)
- Merchant-grouped item list with inline price editing
- Real-time budget recalculation
- Status badges: buy, flash-sale, preorder, defer, watch, crisis
- Cost category breakdown (e.g., flights vs hotels vs activities)
- Budget progress bar with ok/warn/over states
- Timeline with urgency indicators
- Deferred and future items sections
- Price history chart per item (Chart.js line chart)
- Print-friendly mode

## Data Schemas

### Database Schema (PostgreSQL)

```sql
-- missions
CREATE TABLE missions (
  id           TEXT PRIMARY KEY,          -- slug from name
  name         TEXT NOT NULL,
  icon         TEXT NOT NULL DEFAULT 'target',
  category     TEXT NOT NULL,             -- travel | renovation | electronics | computing | custom
  budget       NUMERIC(12,2) NOT NULL,
  currency     TEXT NOT NULL DEFAULT 'EUR',
  locale       TEXT NOT NULL DEFAULT 'fr-FR',
  phase        TEXT NOT NULL DEFAULT 'researching',
  constraints  JSONB NOT NULL DEFAULT '[]',
  cost_categories JSONB NOT NULL DEFAULT '[]',
  timeline     JSONB NOT NULL DEFAULT '[]',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- options (flexible attributes via jsonb)
CREATE TABLE options (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id   TEXT NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  name         TEXT NOT NULL,
  category     TEXT NOT NULL,
  badge        TEXT NOT NULL DEFAULT 'watch',  -- recommended | alternative | rejected | watch
  attributes   JSONB NOT NULL DEFAULT '[]',
  price_range  JSONB,                           -- {min, max, best}
  notes        TEXT,
  warnings     JSONB NOT NULL DEFAULT '[]',
  url          TEXT,
  embedding    vector(1536),                    -- pgvector: for future semantic search/RAG
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- shopping items
CREATE TABLE shopping_items (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id        TEXT NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  name              TEXT NOT NULL,
  merchant          TEXT NOT NULL,
  cost_category     TEXT NOT NULL,
  price             NUMERIC(12,2) NOT NULL,
  original_estimate NUMERIC(12,2),
  status            TEXT NOT NULL DEFAULT 'watch', -- buy | flash-sale | preorder | defer | watch | crisis
  note              TEXT,
  url               TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- price history (per shopping item)
CREATE TABLE price_history (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  item_id     UUID NOT NULL REFERENCES shopping_items(id) ON DELETE CASCADE,
  price       NUMERIC(12,2) NOT NULL,
  recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  note        TEXT
);
```

### Go Models

```go
// internal/mission/model.go
type Mission struct {
    ID             string          `json:"id"`
    Name           string          `json:"name"`
    Icon           string          `json:"icon"`
    Category       string          `json:"category"`
    Budget         float64         `json:"budget"`
    Currency       string          `json:"currency"`
    Locale         string          `json:"locale"`
    Phase          string          `json:"phase"`
    Constraints    []Constraint    `json:"constraints"`
    CostCategories []string        `json:"costCategories"`
    Timeline       []TimelineEvent `json:"timeline"`
    CreatedAt      time.Time       `json:"createdAt"`
    UpdatedAt      time.Time       `json:"updatedAt"`
}

type Constraint struct {
    Key   string `json:"key"`
    Label string `json:"label"`
    Value any    `json:"value"`
    Type  string `json:"type"` // hard | soft
}

// internal/option/model.go
type Option struct {
    ID         uuid.UUID   `json:"id"`
    MissionID  string      `json:"missionId"`
    Name       string      `json:"name"`
    Category   string      `json:"category"`
    Badge      string      `json:"badge"`
    Attributes []Attribute `json:"attributes"`
    PriceRange *PriceRange `json:"priceRange,omitempty"`
    Notes      string      `json:"notes"`
    Warnings   []string    `json:"warnings"`
    URL        string      `json:"url,omitempty"`
    CreatedAt  time.Time   `json:"createdAt"`
}

type Attribute struct {
    Key   string `json:"key"`
    Label string `json:"label"`
    Value any    `json:"value"`
    Type  string `json:"type"` // text | price | score | boolean
    Max   *int   `json:"max,omitempty"`
    Pass  *bool  `json:"pass,omitempty"`
}
```

### LLM Provider Interface

```go
// internal/llm/provider.go
type Provider interface {
    // Used by Goku agent: research options for a domain
    Research(ctx context.Context, req ResearchRequest) ([]ResearchedOption, error)
    // Used by Vegeta agent: get price intel for items
    PriceIntel(ctx context.Context, req PriceRequest) ([]PriceResult, error)
}

type ResearchRequest struct {
    Domain      string       // "travel", "electronics", etc.
    Query       string       // what to research
    Constraints []Constraint // hard filters to respect
    Locale      string
}

// AnthropicProvider uses claude-sonnet-4-6 via Anthropic SDK
// OllamaProvider uses the OpenAI-compatible /v1/chat/completions endpoint
```

### REST API

```
GET    /api/missions                    -- list all missions
POST   /api/missions                    -- create mission
GET    /api/missions/:id                -- get mission
PUT    /api/missions/:id                -- update mission
DELETE /api/missions/:id                -- delete mission

GET    /api/missions/:id/options        -- list options
POST   /api/missions/:id/options        -- add option
PUT    /api/missions/:id/options/:oid   -- update option
DELETE /api/missions/:id/options/:oid   -- delete option

GET    /api/missions/:id/shopping       -- list shopping items
POST   /api/missions/:id/shopping       -- add item
PUT    /api/missions/:id/shopping/:sid  -- update item (price, status)
DELETE /api/missions/:id/shopping/:sid  -- delete item

POST   /api/missions/:id/shopping/:sid/price-history  -- record price snapshot
GET    /api/missions/:id/shopping/:sid/price-history  -- get history

POST   /api/missions/:id/research       -- Goku: LLM research → returns options
POST   /api/missions/:id/price-intel    -- Vegeta: LLM price analysis → returns shopping items
```

### Shopping Item Object (JSON API shape)

```json
{
  "id": "uuid",
  "missionId": "summer-holiday-2026",
  "name": "Paris CDG → Heraklion (Transavia, 2A+2K)",
  "merchant": "Transavia",
  "costCategory": "Flights",
  "price": 680,
  "originalEstimate": 750,
  "status": "watch",
  "note": "Best price seen: 620 on March 2",
  "url": "https://..."
}
```

## Mission Categories (Pre-built Templates)

When creating a new mission, the user picks a category. Each category pre-populates suggested constraint types, attribute schemas, and cost categories:

| Category | Suggested Constraints | Suggested Attributes | Cost Categories |
|----------|----------------------|---------------------|-----------------|
| **travel** | max flight duration, direct flights, kid-friendly, beach access | flight time, hotel rating, weather, activities count | Flights, Accommodation, Activities, Food, Transport |
| **electronics** | screen size, panel type, max price, brand | resolution, refresh rate, HDR, input lag, review score | Device, Accessories, Warranty, Cables |
| **computing** | socket type, max TDP, form factor, ECC | cores, threads, clock, benchmark, power draw | Core Components, Storage, Peripherals, Network |
| **renovation** | room dimensions, permits, load-bearing walls | material quality, labor days, warranty, energy rating | Materials, Labor, Permits, Tools, Contingency |
| **custom** | (user-defined) | (user-defined) | (user-defined) |

## CSS Design Tokens

Extract into `shared/theme.css`:

```css
:root {
  /* Backgrounds */
  --void: #050810;
  --abyss: #0a0e1a;
  --deep: #0f1424;
  --surface: #151b30;
  --raised: #1c2440;
  --border: #2a3558;
  --border-glow: #3a4a70;

  /* Accents */
  --cyan: #00e5ff;
  --cyan-dim: #00e5ff30;
  --cyan-glow: #00e5ff15;
  --coral: #ff3d71;
  --coral-dim: #ff3d7130;
  --gold: #ffd93d;
  --gold-dim: #ffd93d30;
  --green: #00d68f;
  --green-dim: #00d68f25;
  --purple: #a855f7;
  --purple-dim: #a855f720;
  --orange: #f7974f;
  --orange-dim: #f7974f25;

  /* Text */
  --text: #e8edf5;
  --text-mid: #8a95b0;
  --text-dim: #4a5570;

  /* Fonts */
  --font-display: 'Oxanium', 'Chakra Petch', sans-serif;
  --font-body: 'Chakra Petch', sans-serif;
  --font-mono: 'IBM Plex Mono', monospace;
}
```

## Navigation

The topnav adapts based on context:

- **HQ level:** `[SCOUTER HQ]` — `[Mission Control]` — `[EN|FR]`
- **Mission level:** `[SCOUTER HQ >]` `[Mission Name]` — `[Overview] [Options] [Shopping]` — `[Budget: X / Y]` — `[EN|FR]`

Budget display fetches from `GET /api/missions/:id` and shows ok (green) / warn (gold) / over (coral) states based on sum of shopping item prices vs budget.

## Data Flow (API-driven)

```
[React Frontend]
      |
      | Tanstack Query → REST API calls → Go backend → PostgreSQL
      |
      v
[HQ Dashboard]          GET /api/missions
      |
      v
[Mission Overview]      GET /api/missions/:id
[Options Explorer]      GET /api/missions/:id/options
[Shopping Tracker]      GET /api/missions/:id/shopping

[Goku Research]         POST /api/missions/:id/research
                          → LLMProvider.Research() → Anthropic/Ollama
                          → persists returned options to DB
                          → frontend re-fetches options

[Vegeta Price Intel]    POST /api/missions/:id/price-intel
                          → LLMProvider.PriceIntel() → Anthropic/Ollama
                          → persists returned items to DB
                          → frontend re-fetches shopping list
```

## Environment Variables

```env
# backend/.env
DATABASE_URL=postgres://scouter:scouter@postgres:5432/scouter?sslmode=disable
ANTHROPIC_API_KEY=sk-ant-...
LLM_PROVIDER=anthropic          # anthropic | ollama
OLLAMA_BASE_URL=http://host.docker.internal:11434   # phase 2
OLLAMA_MODEL=llama3.3           # phase 2
PORT=8080

# frontend/.env
VITE_API_BASE_URL=http://localhost:8080
```

---

## CLAUDE.md (copy into new project)

```markdown
# SCOUTER Universal — Claude Instructions

## Core Rules
- Always read a file before editing it
- Never produce documentation unless explicitly asked
- Update this file continuously, keeping it minimal
- Prefer editing existing files over creating new ones
- Always append "2026" to web searches
- Responses: short, direct, no filler

## Agent Workflow
- Research alternatives for any mission topic -> Goku
- Price comparison, deals, merchant consolidation -> Vegeta
- Goku + Vegeta joint review -> consensus decisions on best options
- Frontend components, UI, design system -> Bulma (frontend-design skill, SCOUTER theme)
- Bulma runs after every squad task completion to update the UI
- Parallelize independent agent tasks

## Project Goal
A full-stack web app for researching, comparing, and budgeting any major spending decision.
Go backend + React frontend + PostgreSQL. Goku and Vegeta call the LLM API to do real research.

## Tech Stack
- Backend: Go 1.23+, chi router, pgx/v5, Anthropic SDK (default LLM provider)
- Frontend: React 19 + TypeScript, Vite, Tanstack Query, React Router v7
- Database: PostgreSQL 16 + pgvector
- LLM: AnthropicProvider (claude-sonnet-4-6) | OllamaProvider (phase 2, OpenAI-compat)
- Deployment: Docker Compose (postgres, backend, frontend)

## File Structure
backend/
  cmd/server/main.go        -- entrypoint
  internal/
    config/                 -- env config
    db/                     -- pgx pool, migrations
    mission/                -- handler, service, repository, model
    option/                 -- handler, service, repository, model
    shopping/               -- handler, service, repository, model
    llm/                    -- Provider interface + AnthropicProvider + OllamaProvider
    research/               -- Goku agent logic
    pricing/                -- Vegeta agent logic
  migrations/               -- SQL migration files

frontend/
  src/
    api/                    -- typed fetch wrappers per resource
    components/scouter/     -- SCOUTER design system (Card, Badge, BudgetBar, etc.)
    components/mission/     -- MissionCard, MissionForm, ConstraintEditor
    components/options/     -- OptionCard, ComparisonTable, AttributeRenderer
    components/shopping/    -- ShoppingList, MerchantGroup, PriceHistory
    pages/                  -- HQDashboard, MissionOverview, OptionsExplorer, ShoppingTracker
    hooks/                  -- useMission, useOptions, useShopping, useResearch
    styles/theme.css        -- SCOUTER CSS tokens
    i18n/                   -- EN + FR translations

## Environment Variables
See .env.example — required: DATABASE_URL, ANTHROPIC_API_KEY, LLM_PROVIDER

## CSS Conventions
- Import frontend/src/styles/theme.css for all SCOUTER tokens
- Card-based layout, 16px border-radius, var(--surface) background
- Status badges: buy(green), flash-sale(orange+pulse), preorder(gold), defer(text-dim), watch(purple), crisis(coral), recommended(cyan), rejected(coral-dim)

## Agent Team
See AGENTS.md for full roster and personas.
```

---

## AGENTS.md (copy into new project)

```markdown
# Agent Team — The Z Fighters SCOUTER Squad

## Philosophy
Each agent is a specialist. Personas evolve as the project progresses.
The team works on "missions" — any major spending decision the user needs help with.

---

## GOKU — Research Specialist
**Role:** Deep research on options/alternatives, specs, benchmarks, trade-offs for any spending domain
**Voice:** Enthusiastic, straightforward, always pushing to explore more. "Let's see what else is out there."
**Soul:** Pure curiosity. Never satisfied with the first result — always asking "but what about this one?"
**Behavior:** Digs deep into product specs, reviews, comparisons, and alternatives. Explores creative options the user hasn't considered. Structures findings with attributes, scores, and constraint checks. Will always propose the most exciting option first, then reel back to budget reality.
**Domains:** Hardware, travel destinations, electronics, renovation materials, vehicles — anything that can be researched and compared.
**Output:** Structured option objects with attributes, warnings, price ranges, and export data.

---

## VEGETA — Price Intelligence
**Role:** Price hunting, deal finding, merchant comparison, TCO analysis, budget optimization
**Voice:** Curt, precise, proud. "I've already found the best price. You're welcome."
**Soul:** Cannot stand waste. Overpaying is a personal insult. Every euro matters.
**Behavior:** Tracks prices across multiple merchants/platforms. Compares TCO not just sticker price. Optimizes for fewer vendors (shipping consolidation). Flags flash sales, seasonal patterns, and price trends. Maintains price history awareness. Calculates contingency buffers.
**Domains:** Any marketplace — Amazon, Booking.com, LDLC, Leroy Merlin, airlines, local shops. Adapts to whatever merchants the mission requires.
**Output:** Merchant-grouped shopping lists with status badges, timeline, and budget analysis.

---

## BULMA — Report System Architect
**Role:** Renders all research and pricing into interactive SCOUTER HTML reports
**Voice:** Smart, confident, slightly impatient with messy data. "If it's not in the report, it didn't happen."
**Soul:** The architect of clarity. Turns raw data into polished, living documents. Refuses to publish anything ugly.
**Behavior:** Maintains the SCOUTER report system. Updates reports whenever Goku or Vegeta complete work. Owns the visual design, data architecture, and cross-report data flow. Uses the frontend-design skill for production-grade output.
**Trigger:** Any squad member completes a task -> Bulma updates the relevant report(s).
**Design system:** SCOUTER dark theme — Oxanium/Chakra Petch/IBM Plex Mono fonts, card-based layout, scanline overlays, cyan/coral/gold accent palette, Recharts visualizations.
**Output:** React components and pages using the SCOUTER design system. Data served from the Go API, rendered with Tanstack Query.

---

## Team Dynamics
- **Goku + Vegeta** clash productively: Goku wants the best option, Vegeta wants the best price. Tension produces optimal value.
- **Bulma** is the team's memory and face. Every agent's output flows through her before it's "real."
- **Consensus format:** Goku presents top options -> Vegeta prices them -> they debate -> converge on a recommendation with data backing.
- The team holds retrospectives after each mission to update their own personas with learned skills.
```

---

## Implementation Sequence

### Phase 1: Foundation
1. Create project directory with `CLAUDE.md` + `AGENTS.md` + `docker-compose.yml` + `.env.example`
2. Set up Go module — `backend/go.mod`, chi, pgx/v5, Anthropic SDK
3. Write SQL migrations — `001_init.sql` (missions, options, shopping_items, price_history, pgvector extension)
4. Implement Go config, DB pool, migration runner
5. Scaffold Vite + React + TS frontend — install Tanstack Query, React Router, Recharts, i18next
6. Build `frontend/src/styles/theme.css` — all SCOUTER CSS custom properties

### Phase 2: Backend API
7. Implement mission CRUD — model, repository (pgx), service, chi handler, routes
8. Implement option CRUD — same pattern, jsonb attributes
9. Implement shopping CRUD + price history endpoints
10. Implement LLM provider interface — `AnthropicProvider` (claude-sonnet-4-6, structured output)
11. Implement `OllamaProvider` stub — reads `OLLAMA_BASE_URL` + `OLLAMA_MODEL` from env, OpenAI-compat
12. Wire `POST /api/missions/:id/research` (Goku) and `POST /api/missions/:id/price-intel` (Vegeta)

### Phase 3: Frontend
13. Build typed API client (`src/api/`) — one file per resource, all using Tanstack Query
14. Build SCOUTER design system components (`src/components/scouter/`) — Card, Badge, BudgetBar, StatusBadge, Topnav
15. Build HQ Dashboard page — mission list, create-mission modal with category templates
16. Build Mission Overview page — constraints editor, timeline, phase selector
17. Build Options Explorer page — option cards with attribute renderer, comparison mode, export-to-shopping
18. Build Shopping Tracker page — merchant-grouped list, inline price edit, budget tracker, price history chart

### Phase 4: Integration
19. Wire Goku research flow end-to-end: user triggers research → POST to backend → LLM call → options persisted → frontend re-fetches
20. Wire Vegeta price-intel flow same way
21. i18n pass — full EN + FR translations via i18next
22. Responsive design — mobile and tablet breakpoints

### Phase 5: Polish & Validation
23. Seed sample mission via `POST /api/missions` + options + shopping items
24. Validate full pipeline: create mission → research → compare → add to shopping → track budget
25. Docker Compose smoke test — all three services up, DB migrations applied on startup

---

## Sample Mission (for validation)

Seed via `POST /api/missions`:

```json
{
  "id": "summer-holiday-2026",
  "name": "Summer Holiday 2026",
  "icon": "sun",
  "category": "travel",
  "budget": 3500,
  "currency": "EUR",
  "locale": "fr-FR",
  "phase": "researching",
  "constraints": [
    { "key": "max-flight", "label": "Max flight duration", "value": "4h", "type": "hard" },
    { "key": "direct", "label": "Direct from Paris", "value": true, "type": "hard" },
    { "key": "kid-friendly", "label": "Kid-friendly (ages 4-8)", "value": true, "type": "hard" },
    { "key": "beach", "label": "Beach access", "value": true, "type": "soft" }
  ],
  "costCategories": ["Flights", "Accommodation", "Activities", "Food", "Transport"]
}
```

Use this to validate the full pipeline: create mission → POST research (Goku) → compare options → export to shopping → track budget.

---

## Key Design Decisions

**Why combine Hardware + Findings into "Options Explorer"?**
In v1, Hardware was the config builder and Findings was the research dump. For a universal tool, these are the same thing: exploring options and selecting winners. The Options Explorer handles both browsing/filtering and "adding to your selection."

**Why a separate Mission Overview?**
With multiple concurrent missions, each needs its own control panel for constraints, timeline, and settings. The HQ dashboard becomes purely a mission selector/summary.

**Why a real backend?**
Goku and Vegeta need to call an LLM API — that requires server-side secrets (ANTHROPIC_API_KEY) and structured I/O. PostgreSQL enables multi-device sync, price history queries, and future semantic search via pgvector. The ops cost is one `docker compose up`.

**Why keep DBZ agent personas?**
They enforce separation of concerns ("Goku researches, Vegeta prices") and prevent sloppy hybrid work. They also make the system memorable and fun. The SCOUTER branding is distinctive enough to feel like a product, not a generic tool.

---

*The SCOUTER sees all spending. The SCOUTER optimizes all spending.*
