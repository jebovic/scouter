# Functional Overview

> SCOUTER Universal is a **personal spending intelligence tool** that guides you from the first question — *"Should I buy this?"* — all the way through research, price tracking, purchase, and post-purchase reflection.

## The Problem SCOUTER Solves

Making a major purchase — a laptop, a car, an appliance — is stressful. You face:

- Too many options with no clear comparison
- Price uncertainty (is this a good deal today?)
- Budget discipline (am I spending within my limits?)
- Decision regret (did I make the right call?)

SCOUTER automates the research, tracks prices over time, scores deals in real time, and helps you learn from past purchases.

---

## Mission Lifecycle

![Mission Lifecycle](../assets/mission-lifecycle.svg)

A **Mission** is SCOUTER's core unit — one purchase goal with a budget, category, and constraints.

| Step | What Happens |
|------|-------------|
| **1. Create** | Name the mission, set budget, choose category, define constraints (max weight, min RAM, etc.) |
| **2. Research** | ResearchAgent (LLM) discovers and ranks options based on your requirements |
| **3. Pricing** | PricingAgent hunts real prices, calculates Total Cost of Ownership, scores each deal |
| **4. Compare** | Options Explorer — radar charts, constraint checker, comparison table side-by-side |
| **5. Track** | Shopping Tracker — price history, budget burn, deal alerts, merchant comparison |
| **6. Purchase** | Record the purchase, capture lessons learned, mission advances to "done" |
| **7. Scorecard** | Efficiency grade (A/B/C/D), analytics, spending history |

---

## Feature Catalog

### Core Intelligence

| Feature | Description |
|---------|-------------|
| **ResearchAgent** | LLM discovers options using tool use, ranks by fit to your constraints |
| **PricingAgent** | Hunts prices across merchants, calculates TCO, scores deal quality |
| **Deal Scoring** | Real-time deal score combining price trend, urgency, budget fit |
| **Price Alerts** | Background scheduler checks prices on active missions, fires notifications |
| **Semantic Search** | pgvector cosine similarity search across all options (300ms debounce) |
| **Similar Options** | Find options similar to one you're viewing |

### Budget & Finance

| Feature | Description |
|---------|-------------|
| **Budget Bar** | Visual budget burn progress per mission |
| **Envelope Budgeting** | Allocate monthly envelopes to spending categories |
| **Price Forecast** | LLM-powered price trend prediction |
| **French Benchmark** | Compare price to French market median (bon prix / prix moyen / au-dessus) |
| **ROI Calculator** | Calculate return on investment for the purchase |
| **Inflation Tracker** | Track price changes relative to inflation |
| **Cash-Back Tracker** | Monitor cash-back rewards and loyalty points |

### Shopping Intelligence

| Feature | Description |
|---------|-------------|
| **Quantity Optimizer** | Discount tiers [1, 2, 3, 5, 10] with FNV-32a discount curve |
| **Timeline Planner** | 4-week budget distribution plan with French promo hints |
| **Wishlist Prioritizer** | Rank wishlist items by urgency + trend + budget fit |
| **Deal Calendar** | Flash sales and promotional calendar |
| **Coupon Finder** | Agent discovers applicable coupons |
| **Negotiation Tips** | LLM-powered negotiation advice for each merchant |
| **Merchant Recommender** | Rank merchants by price + reliability + shipping |

### Collaboration

| Feature | Description |
|---------|-------------|
| **Mission Sharing** | Share tokens for read-only public access |
| **Invites** | Invite collaborators to a mission |
| **Voting** | Collaborators vote on options |
| **Comments** | Threaded comments on options |
| **Public Wishlist** | Shareable wishlist page |

### Analytics & Reporting

| Feature | Description |
|---------|-------------|
| **Scorecard** | Efficiency grade A/B/C/D from 4 DB queries, 30min cache |
| **Stats Page** | Total spending + category breakdown |
| **Performance Page** | Mission efficiency metrics over time |
| **History Page** | Completed missions + purchase timeline |
| **Insights Page** | Price insights and trends |
| **Weekly Digest** | Email-style digest of price changes |
| **Spending Persona** | LLM-powered spending personality analysis |

### Accessibility & UX

| Feature | Description |
|---------|-------------|
| **i18n** | Full EN + FR translations (react-i18next) |
| **Keyboard Shortcuts** | `N` new mission · `R` research · `P` pricing |
| **PWA** | Installable, offline-capable (service worker) |
| **Onboarding** | 3-step overlay, localStorage dismissed |
| **Responsive** | Mobile-first, 640px and 1024px breakpoints |
| **Skeleton Loading** | Card/row/chart skeleton variants |
| **Empty States** | Icon + title + description + CTA action |
| **Dark Theme** | SCOUTER design system with CSS custom properties |

---

## Mission Templates

15 built-in templates compiled into the binary (GET /api/templates):

| Category | Templates |
|----------|-----------|
| Electronics | Laptop, Smartphone, TV, Camera |
| Home | Appliance, Furniture, Renovation |
| Transport | Vehicle, Bicycle, Electric Scooter |
| Finance | Investment, Insurance |
| Travel | Vacation planning |
| Health | Equipment, Subscription |

Each template pre-fills budget, category, and suggested constraints.

---

## Status System

Options are tagged with deal status displayed as colored badges:

| Status | Color | Meaning |
|--------|-------|---------|
| `buy` | Green | Strong buy recommendation |
| `flash-sale` | Orange + pulse | Time-limited offer |
| `preorder` | Gold | Pre-order available |
| `recommended` | Cyan | Recommended option |
| `watch` | Purple | Watching for price drop |
| `defer` | Text-dim | Defer — not the right time |
| `crisis` | Coral | Price spike / crisis |
| `rejected` | Coral-dim | Rejected option |
