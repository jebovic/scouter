# Mission & Option Management — Design Spec (Part 1)

**Date:** 2026-03-16
**Status:** Approved
**Scope:** UI management actions for missions and options, plus pin-as-shortlist workflow fix

---

## Problem Statement

Three friction points block users from managing missions effectively:

1. **No delete or visible archive** — missions can only be archived via a hidden swipe gesture; there is no delete button anywhere in the UI
2. **Nothing is editable after creation** — mission fields (name, budget, constraints) and option fields (name, notes, price range) are frozen once set
3. **Pin has no payoff** — options can be pinned but the pin does nothing beyond a visual highlight; the "Clear Pinned" button permanently deletes pinned options, which is the opposite of what a user expects

---

## Decisions

- **Editing pattern**: Modal form (not inline edit) — consistent with existing form UX in the app
- **Delete semantics**: Both archive (soft, reversible) and hard delete are exposed; same pattern for missions and options
- **Pin semantics**: Pin = shortlist candidate; the shortlist surfaces in the buying phase to pre-fill the purchase form
- **Approach**: Action bar on mission overview + 3-dot menu on option cards (proportional to the importance of each entity)

---

## Design

### 1. Mission Management Actions

#### Mission Overview (action bar in header)

A small action bar appears in the MissionOverview page header, placed to the right of the existing phase switcher buttons. It contains three actions:

| Action | Behavior |
|--------|----------|
| **Edit Mission** | Opens `MissionEditModal`, pre-filled with current values. Calls `PATCH /api/missions/{slug}` on save. |
| **Archive** | Single confirmation dialog. Soft-deletes by calling `POST /api/missions/{id}/archive`. Mission disappears from the main dashboard; accessible via the new "Archived" toggle on HQ Dashboard. |
| **Delete** | Destructive confirmation dialog: "This will permanently delete the mission and all its options. Type the mission name to confirm." Calls `DELETE /api/missions/{slug}`. On success, navigates back to `/`. |

**MissionEditModal fields** (reuses existing `MissionForm` component with `initialValues` + adds `ConstraintEditor`):
- Name (text input)
- Icon (emoji picker)
- Category (dropdown: travel / electronics / computing / renovation / custom)
- Budget (numeric input)
- Currency (dropdown: USD / EUR / GBP / CAD / AUD / JPY)
- Constraints (via existing `ConstraintEditor` component)
- Envelope (via existing `EnvelopeSelector` component — already in `MissionForm`, no change needed)

**Error handling:**
- API failure on save → show inline error message in modal, keep modal open
- API failure on archive/delete → show a toast error; do not navigate away

#### HQ Dashboard (mission cards)

Mission cards gain a `⋯` overflow menu (top-right). Menu items: **Archive** and **Delete** (same behavior as above). **Edit is not in this menu** — to edit, the user navigates into the mission via the card's existing title link, then uses the action bar there.

**Archived missions toggle:** A new filter button ("Show archived") appears in the HQ Dashboard header. When active, archived missions are included in the list with a visual "archived" badge. Each archived mission's `⋯` menu gains an **Unarchive** item (replacing the Archive item) that calls `POST /api/missions/{id}/unarchive` and restores the mission to the active list.

---

### 2. Option Management Actions

#### Options Explorer (option cards)

Option cards gain a `⋯` overflow menu alongside the existing pin/reject buttons. Menu items:

| Action | Behavior |
|--------|----------|
| **Edit** | Opens `OptionEditModal` pre-filled with current values. Calls `PUT /api/missions/{missionId}/options/{optionId}` on save. |
| **Delete** | Confirmation dialog: "Remove this option permanently?" Calls `DELETE /api/missions/{missionId}/options/{optionId}`. On success, removes option from the list. |

**OptionEditModal fields:**
- Name (text input)
- Badge (dropdown): `recommended` / `alternative` / `watch` / `rejected`
- Price range: Min and Max (numeric inputs, currency inherited from mission)
- Notes (textarea)
- Warnings (textarea)
- Attributes: **read-only display** — shown for context, not editable (Part 2 scope)

**Error handling:**
- API failure on save → inline error message in modal, keep modal open
- API failure on delete → toast error; option remains visible

#### "Clear Pinned" → "Unpin All"

The existing "Clear Pinned" button is renamed to **"Unpin All"**. Behavior change: calls `PATCH /api/missions/{missionId}/options/{optionId}/pin` in **parallel** for all currently pinned options to toggle them off. All options remain in the explorer, just unpinned.

**Partial failure:** if some pin toggles fail, show a toast: "Could not unpin {n} option(s). Please try again." Leave failed options in their current pinned state.

---

### 3. Pin = Shortlist (Buying Phase)

#### Shortlist Panel in MissionOverview

When the mission phase is **buying**, a "Your Shortlist" panel appears above the purchase form in MissionOverview. It lists all pinned options as compact cards showing: name, price range, badge.

Each shortlist card has a **Select** button. Clicking it pre-fills the purchase form fields:
- Option name → purchase item name field
- Price range midpoint → suggested price field

**If a purchase record already exists** when "Select" is clicked: show a confirmation prompt — "A purchase is already recorded. Replace with this option?" — with Confirm / Cancel. On confirm, overwrite the purchase form fields.

**If no options are pinned** when entering buying phase, the panel renders an empty state prompt:
*"Go back to Options and pin your top picks to build your shortlist."*
With a link to the Options Explorer for this mission.

#### Pin Behavior (unchanged)

The pin toggle on option cards works the same as today. The only change is that pinned options now have a clear purpose: they accumulate into the shortlist visible in the buying phase.

---

## Backend Impact

All required backend endpoints already exist. No new endpoints are needed for Part 1:

| Endpoint | Status |
|----------|--------|
| `PATCH /api/missions/{slug}` | ✅ exists |
| `DELETE /api/missions/{slug}` | ✅ exists |
| `POST /api/missions/{id}/archive` | ✅ exists |
| `GET /api/missions?include_archived=true` | ✅ exists |
| `PUT /api/missions/{missionId}/options/{optionId}` | ✅ exists |
| `DELETE /api/missions/{missionId}/options/{optionId}` | ✅ exists |
| `PATCH /api/missions/{missionId}/options/{optionId}/pin` | ✅ exists |

---

## Frontend Impact

### New components
- `MissionActionBar` — edit/archive/delete actions for mission overview header
- `MissionEditModal` — wraps `MissionForm` + `ConstraintEditor` in a modal for editing existing missions
- `OptionEditModal` — form modal for editing option fields
- `ShortlistPanel` — buying-phase panel showing pinned options with Select button

### Modified components
- `MissionCard` (HQ Dashboard) — add `⋯` menu with archive/delete; add archived badge variant
- `HQDashboard` — add "Show archived" toggle that passes `include_archived=true` to missions query
- `OptionCard` — add `⋯` menu with edit/delete; pin button behavior unchanged
- `OptionsExplorer` — rename "Clear Pinned" → "Unpin All"; change behavior to parallel unpin with partial-failure toast
- `MissionOverview` — integrate `MissionActionBar` in header; integrate `ShortlistPanel` above purchase form when phase = buying

### i18n keys (all required in `en.json` and `fr.json`)

```
mission.actions.edit
mission.actions.archive
mission.actions.delete
mission.actions.archiveConfirm
mission.actions.deleteConfirm
mission.actions.deleteTypeToConfirm
mission.actions.showArchived
mission.actions.unarchive
mission.actions.archivedBadge

option.actions.edit
option.actions.delete
option.actions.deleteConfirm
option.actions.unpinAll
option.actions.unpinAllError

options.badge.recommended
options.badge.alternative
options.badge.watch
options.badge.rejected

shortlist.title
shortlist.empty
shortlist.emptyLink
shortlist.select
shortlist.replaceConfirm
```

---

## Out of Scope (Part 2)

- Editing option attributes (the Research Agent-generated structured specs)
- Composable options (building an option from individual components)
- Auto-populating the shopping list from the chosen option's components
