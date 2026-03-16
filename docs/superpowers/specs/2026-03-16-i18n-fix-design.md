# i18n Fix & Lint Guard — Design Spec

**Date:** 2026-03-16
**Status:** Approved
**Scope:** Frontend — 4 pages + ESLint enforcement

---

## Problem

Hardcoded French (and occasionally English) strings appear in the UI despite the app requiring full EN/FR i18n coverage via `react-i18next`. These strings bypass `t()` and are invisible to the translation system, causing language-switching to silently fail for affected text.

**Root causes identified:**
- No automated guard prevents literal strings in JSX
- Development happens in a mixed EN/FR style — strings are written inline and never backfilled into i18n keys
- Common leak patterns: form validation messages, ternary status labels, input placeholders, page titles, empty-state copy, button labels

**Affected files (43 hardcoded strings total):**

| File | Count | Representative examples |
|------|-------|------------------------|
| `CashbackPage.tsx` | ~18 | `"En attente"`, `"Confirmé"`, `"Reçu"`, `"Ajouter"`, `"Supprimer"` |
| `EnvelopesPage.tsx` | ~16 | `"Le nom est requis"`, `"Annuler"`, `"Budget (€)"`, `"Transactions récentes"` |
| `InsightsPage.tsx` | ~10 | `"Analyse en cours…"`, `"Aucune donnée"`, `"Par catégorie"` |
| `LoyaltyPage.tsx` | 2 | Page title and subtitle |

---

## Solution

### Part 1 — ESLint guard (`eslint-plugin-i18next`)

Install `eslint-plugin-i18next` and enable it in the existing ESLint 9 flat config (`eslint.config.js`). The rule `i18next/no-literal-string` flags any JSX text node or JSX string attribute that is not wrapped in a `t()` call.

**Configuration approach:**
- Use `i18next.configs['flat/recommended']` as the base
- Override to `"error"` severity so `npm run lint` fails on violations
- Tune ignore patterns to suppress false positives on: pure numbers, single characters, CSS class strings, `aria-*` attribute values, `data-*` attributes, and `key` props

This runs on every `npm run lint` invocation and in VSCode via the ESLint extension (real-time squiggles).

### Part 2 — Fix pass on 43 occurrences

For each affected file:
1. Ensure `const { t } = useTranslation()` is imported and called
2. Replace every hardcoded string with `t('namespace.key')`
3. Add the key to both `en.json` (English value) and `fr.json` (French value)

**Namespace conventions** (matching existing codebase patterns):
- `cashback.*` — CashbackPage strings
- `envelopes.*` — EnvelopesPage strings
- `insights.*` — InsightsPage strings
- `loyalty.*` — LoyaltyPage strings
- `common.*` — shared strings (e.g., `common.cancel`, `common.save`, `common.loading`, `common.delete`)

**Key i18n entries to add (representative, not exhaustive):**

```json
// en.json additions
{
  "common": {
    "cancel": "Cancel",
    "save": "Save",
    "delete": "Delete",
    "loading": "Loading...",
    "add": "Add",
    "unknown": "Unknown"
  },
  "cashback": {
    "title": "Cashback Tracker",
    "subtitle": "Track your cashback refunds",
    "pending": "Pending",
    "confirmed": "Confirmed",
    "received": "Received",
    "addTitle": "Add cashback",
    "item": "Item",
    "merchant": "Merchant",
    "amount": "Amount (€)",
    "rate": "Rate (%)",
    "status": "Status",
    "myCashbacks": "My cashbacks",
    "none": "No cashback recorded",
    "noneDesc": "Use the form above to log your first refund."
  },
  "envelopes": {
    "nameRequired": "Name is required",
    "invalidBudget": "Invalid budget",
    "newEnvelope": "New envelope",
    "namePlaceholder": "e.g. Groceries, Entertainment…",
    "budget": "Budget (€)",
    "labelRequired": "Label is required",
    "invalidAmount": "Invalid amount",
    "addOperation": "Add transaction",
    "expense": "Expense",
    "topup": "Top-up",
    "label": "Label",
    "labelPlaceholder": "e.g. Supermarket, Netflix…",
    "amount": "Amount (€)",
    "date": "Date",
    "spent": "Spent",
    "remaining": "Remaining",
    "overbudget": "Overbudget",
    "addExpense": "+ Add expense",
    "recentTransactions": "Recent transactions",
    "deleteTransaction": "Delete transaction",
    "title": "Envelope Budget",
    "subtitle": "Manage your budget by category",
    "newEnvelopeBtn": "+ New envelope",
    "globalBudget": "GLOBAL BUDGET",
    "none": "No envelopes",
    "noneDesc": "Create envelopes to organise your budget by category.",
    "createFirst": "Create my first envelope"
  },
  "insights": {
    "loading": "Analysis in progress…",
    "noData": "No data",
    "noDataDesc": "Create missions to see your insights.",
    "title": "Budget Insights",
    "activeMissions": "Active missions",
    "totalBudget": "Total budget",
    "budgetUsage": "Budget usage",
    "recommendations": "Recommendations",
    "byCategory": "By category",
    "allMissions": "All missions"
  },
  "loyalty": {
    "title": "My Loyalty Points",
    "subtitle": "Track your loyalty programmes"
  }
}
```

```json
// fr.json additions (mirror with French values)
{
  "common": {
    "cancel": "Annuler",
    "save": "Enregistrer",
    "delete": "Supprimer",
    "loading": "Chargement...",
    "add": "Ajouter",
    "unknown": "Inconnu"
  },
  "cashback": { ... },
  "envelopes": { ... },
  "insights": { ... },
  "loyalty": {
    "title": "Mes Points Fidélité",
    "subtitle": "Suivi de vos programmes de fidélité"
  }
}
```

---

## Out of Scope

- `de.json` (German stub — not actively supported)
- Backend strings
- Pages not in the 4 identified above
- Pre-commit hooks (ESLint via `npm run lint` is sufficient)

---

## Success Criteria

1. `npm run lint` passes with zero `i18next/no-literal-string` errors after the fix pass
2. All 4 affected pages render correctly in both EN and FR locales
3. Switching language via the app settings shows correct translated text on all fixed pages
4. Any future PR that introduces a hardcoded JSX string will fail `npm run lint`

---

## Implementation Order

1. Install `eslint-plugin-i18next` as devDependency
2. Update `eslint.config.js` to enable the rule
3. Run `npm run lint` to confirm the 43 violations are surfaced
4. Fix `en.json` and `fr.json` (add all missing keys)
5. Fix `InsightsPage.tsx` (smallest, good smoke test)
6. Fix `LoyaltyPage.tsx`
7. Fix `CashbackPage.tsx`
8. Fix `EnvelopesPage.tsx` (largest)
9. Run `npm run lint` again — should be zero violations
10. Run `npm run build` to confirm no type errors
