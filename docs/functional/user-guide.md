# User Guide

Step-by-step guide to getting the most out of SCOUTER Universal.

---

## Quick Start

1. Start SCOUTER with Docker Compose (see [Quick Start](../deployment/quickstart.md))
2. Open `http://localhost:5173`
3. Click **New Mission** (or press `N`)
4. Choose a template or configure from scratch

---

## Creating a Mission

### From a Template

1. On the HQ Dashboard, click **New Mission**
2. Click **Browse Templates** to open the Template Gallery
3. Click a template card to preview it
4. Click **Use Template** — budget, category, and constraints are pre-filled
5. Customize the budget and constraints if needed
6. Click **Create Mission**

### From Scratch

1. Click **New Mission**
2. Fill in:
   - **Name**: e.g. "Work Laptop Upgrade"
   - **Category**: Electronics, Travel, Home…
   - **Budget**: Your maximum spend
   - **Constraints**: Min RAM, max weight, specific OS…
3. Click **Create Mission**

---

## Running Research

1. Open a mission → click **Research** (or press `R` from the mission page)
2. SCOUTER's ResearchAgent calls the LLM with your constraints as tool schema
3. The agent discovers 5–10 options, ranks them, and persists them to your database
4. Options appear in the **Options Explorer** tab

**What the agent does:**
- Builds a structured prompt from your mission constraints
- Uses tool use to get structured option data (name, brand, price, attributes)
- Embeds each option asynchronously via OllamaEmbedder (for semantic search)
- Ranks options by fit to your requirements

---

## Comparing Options

Go to **Options Explorer** (the chart icon in the mission nav):

- **Radar Chart**: Multi-axis comparison (performance, value, reliability, etc.)
- **Comparison Table**: Side-by-side attribute comparison
- **Constraint Checker**: Shows which options violate your constraints (highlighted in red)
- **Status Badge**: Deal quality at a glance
- **Similar Options**: Click any option to find semantically similar alternatives

**Keyboard tips:**
- Use semantic search (`Cmd/Ctrl + K`) to search across all options

---

## Running Pricing

1. From any option card, click **Get Prices** (or press `P` from the mission page)
2. PricingAgent hunts prices across merchants:
   - Fetches current prices
   - Calculates Total Cost of Ownership (TCO)
   - Scores the deal (0–100)
   - Recommends best merchant
3. Prices appear in the **Shopping Tracker** tab

---

## Shopping Tracker

The Shopping Tracker shows per-item deal intelligence:

| Column | Description |
|--------|-------------|
| **Item** | Name + status badge |
| **Price** | Current best price |
| **Deal Score** | 0–100 composite score |
| **Trend** | Price direction (up/flat/down) |
| **Target Price** | Your desired price — edit inline |
| **Budget** | Budget bar showing burn % |

### Timeline Planner
For each item, the Timeline Planner (Phase 172) shows a 4-week distribution:
- Which weeks are best to buy based on status patterns
- French promotional calendar hints (soldes, Black Friday…)

### Quantity Optimizer
For bulk purchases, the Quantity Optimizer (Phase 171) shows:
- Discount tiers for quantities [1, 2, 3, 5, 10]
- Savings at each tier

### French Benchmark
For each mission (Phase 169):
- Market median price from French retailers
- Verdict: **bon prix** / **prix moyen** / **au-dessus du marché**

---

## Semantic Search

Press `Cmd/Ctrl + K` or click the search icon in the top navigation:

1. Start typing — results appear after 2 characters with 300ms debounce
2. Up to 5 results shown in the dropdown
3. Press `Enter` to go to the full search page (`/search`)
4. Search uses pgvector cosine similarity against embedded option text

---

## Notifications

Click the bell icon in the top navigation:

- Price alerts fire when a tracked item drops below your target price
- Background scheduler checks active missions on a schedule
- Mark individual notifications as read, or mark all
- Unread count badge updates every 60 seconds

---

## Purchase & History

When you're ready to buy:

1. Open the mission → click **Record Purchase**
2. Fill in:
   - **Merchant**: Where you bought it
   - **Price Paid**: Actual amount
   - **Date**: Purchase date
   - **Lessons**: What you'd do differently next time
   - **Rating**: 1–5 stars
3. Mission status advances to **done** automatically
4. View in **History Page** (`/history`)

---

## Stats & Analytics

**Stats Page** (`/stats`):
- Total spend across all missions
- Spend by category (bar chart)
- Average deal score
- Missions completed

**Performance Page** (`/performance`):
- Mission efficiency over time
- Scorecard grades (A/B/C/D)
- Trend analysis

**Scorecard** (on each completed mission):
- Efficiency grade A/B/C/D based on:
  - Price vs market median
  - Time to decision
  - Research depth
  - Budget discipline
- Achievements and lesson history

---

## Settings

Settings Page (`/settings`):

| Setting | Options |
|---------|---------|
| **Currency** | EUR, USD, GBP, CAD, CHF… |
| **Locale** | fr-FR, en-US, en-GB… |
| **LLM Provider** | anthropic, ollama, routing |
| **Danger Zone** | Delete all data (two-step confirm) |

---

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `N` | New mission |
| `R` | Run research on current mission |
| `P` | Run pricing on current mission |
| `Cmd/Ctrl + K` | Open semantic search |

---

## Collaboration

To share a mission:

1. Open a mission → click **Share**
2. Copy the share link (public read-only)
3. To invite collaborators: click **Invite** → enter email or username
4. Collaborators can view options and vote
5. Revoke access at any time with **Remove Share**

---

## Wishlist

The Wishlist (`/wishlist`) collects items you're not ready to buy yet:

- Add items with a target price
- Set price alerts (triggers when price drops below target)
- **Prioritized View**: Items ranked by urgency + trend + budget fit
- **Shared Wishlist**: Generate a public share link for gift lists
