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

**Affected files (~43 hardcoded strings total):**

| File | Count | Representative examples |
|------|-------|------------------------|
| `CashbackPage.tsx` | ~18 | `"En attente"`, `"Confirmé"`, `"Reçu"`, `"Ajouter"`, `"Supprimer"` |
| `EnvelopesPage.tsx` | ~16 | `"Le nom est requis"`, `"Annuler"`, `"Budget (€)"`, `"Transactions récentes"` |
| `InsightsPage.tsx` | ~10 | `"Analyse en cours…"`, `"Aucune donnée"`, `"Par catégorie"` |
| `LoyaltyPage.tsx` | 2 | Page title and subtitle |

---

## Solution

### Part 1 — ESLint guard (`eslint-plugin-i18next`)

Install `eslint-plugin-i18next` as a devDependency and enable it in the existing ESLint 9 flat config. The rule `i18next/no-literal-string` flags any JSX text node or JSX string attribute not wrapped in a `t()` call.

**Exact config block to add to `eslint.config.js`:**

```js
import i18next from 'eslint-plugin-i18next'

// Add inside the defineConfig([...]) array, after existing rules:
{
  // files must come AFTER the spread to avoid being silently overridden
  ...i18next.configs['flat/recommended'],
  files: ['**/*.{ts,tsx}'],
  rules: {
    ...i18next.configs['flat/recommended'].rules,
    'i18next/no-literal-string': ['error', {
      mode: 'jsx-only',
      'jsx-attributes': {
        include: ['title', 'placeholder', 'aria-label', 'alt'],
      },
      // ignore: pure numbers, alphanumeric tokens (incl. spaces for placeholders like "MacBook Pro"), ALL_CAPS constants
      ignore: [/^\d+(\.\d+)?$/, /^[a-zA-Z0-9_\-\.\/ ]+$/, /^[A-Z_]+$/],
    }],
  },
},
```

This runs on every `npm run lint` invocation and in VSCode via the ESLint extension (real-time squiggles).

### Part 2 — Fix pass on ~43 occurrences

For each affected file:
1. Ensure `const { t } = useTranslation()` is imported and called (**note: `InsightsPage.tsx` already has this — do not duplicate**)
2. Replace every hardcoded string with `t('namespace.key')`
3. Use interpolation for dynamic strings: `t('cashback.deleteAriaLabel', { name: entry.itemName })`
4. Add every new key to both `en.json` and `fr.json`

---

## i18n Keys to Add

### Existing keys — do NOT re-add these

These already exist in `en.json` / `fr.json` and must be **reused**, not duplicated:

| Key | EN value | Note |
|-----|----------|------|
| `common.cancel` | "Cancel" | already exists |
| `common.save` | "Save" | already exists |
| `common.delete` | "Delete" | already exists |
| `common.loading` | "Loading..." | already exists |
| `common.add` | "Add" | already exists |
| `envelopes.empty` | "No envelopes" | already exists — maps to `"Aucune enveloppe"` in FR |
| `envelopes.emptyDesc` | (already exists) | maps to `"Créez des enveloppes…"` in FR |

### New keys — add to both `en.json` and `fr.json`

#### `common` additions

| Key | EN | FR |
|-----|----|----|
| `common.unknown` | "Unknown" | "Inconnu" |

#### `cashback` namespace (currently empty)

| Key | EN | FR |
|-----|----|----|
| `cashback.title` | "Cashback Tracker" | "Cashback Tracker" |
| `cashback.subtitle` | "Track your cashback refunds" | "Suivez vos remboursements cashback" |
| `cashback.pending` | "Pending" | "En attente" |
| `cashback.confirmed` | "Confirmed" | "Confirmé" |
| `cashback.received` | "Received" | "Reçu" |
| `cashback.addTitle` | "+ Add cashback" | "+ Ajouter un cashback" |
| `cashback.item` | "Item" | "Article" |
| `cashback.merchant` | "Merchant" | "Marchand" |
| `cashback.amount` | "Amount (€)" | "Montant (€)" |
| `cashback.rate` | "Rate (%)" | "Taux (%)" |
| `cashback.status` | "Status" | "Statut" |
| `cashback.notes` | "Notes" | "Notes" |
| `cashback.optional` | "Optional" | "Optionnel" |
| `cashback.adding` | "Adding..." | "Ajout..." |
| `cashback.myCashbacks` | "My cashbacks" | "Mes cashbacks" |
| `cashback.none` | "No cashback recorded" | "Aucun cashback enregistré" |
| `cashback.noneDesc` | "Use the form above to log your first refund." | "Utilisez le formulaire ci-dessus pour logger votre premier remboursement." |
| `cashback.delete` | "Delete" | "Supprimer" |
| `cashback.deleteAriaLabel` | "Delete {{name}}" | "Supprimer {{name}}" |

#### `envelopes` additions (namespace partially exists — keys below are all NEW)

| Key | EN | FR |
|-----|----|----|
| `envelopes.title` | "Envelope Budget" | "Budget Enveloppes" |
| `envelopes.pageSubtitle` | "Manage your budget by category" | "Gérez votre budget par catégorie" |
| `envelopes.newBtn2` | "+ New envelope" | "+ Nouvelle enveloppe" |
| `envelopes.globalBudget` | "GLOBAL BUDGET" | "BUDGET GLOBAL" |
| `envelopes.nameRequired` | "Name is required" | "Le nom est requis" |
| `envelopes.invalidBudget` | "Invalid budget" | "Budget invalide" |
| `envelopes.newEnvelope` | "New envelope" | "Nouvelle enveloppe" |
| `envelopes.namePlaceholder` | "e.g. Groceries, Entertainment…" | "ex: Alimentation, Loisirs…" |
| `envelopes.budget` | "Budget (€)" | "Budget (€)" |
| `envelopes.labelRequired` | "Label is required" | "Libellé requis" |
| `envelopes.invalidAmount` | "Invalid amount" | "Montant invalide" |
| `envelopes.addOperation` | "Add transaction" | "Ajouter une opération" |
| `envelopes.expense` | "Expense" | "Dépense" |
| `envelopes.topup` | "Top-up" | "Rechargement" |
| `envelopes.label` | "Label" | "Libellé" |
| `envelopes.labelPlaceholder` | "e.g. Supermarket, Netflix…" | "ex: Supermarché, Netflix…" |
| `envelopes.amount` | "Amount (€)" | "Montant (€)" |
| `envelopes.date` | "Date" | "Date" |
| `envelopes.spent` | "Spent" | "Dépensé" |
| `envelopes.remaining` | "Remaining" | "Restant" |
| `envelopes.overbudget` | "Overbudget" | "Dépassé" |
| `envelopes.addExpense` | "+ Add expense" | "+ Ajouter une dépense" |
| `envelopes.recentTransactions` | "Recent transactions" | "Transactions récentes" |
| `envelopes.deleteTransaction` | "Delete transaction" | "Supprimer la transaction" |
| `envelopes.createFirst` | "Create my first envelope" | "Créer ma première enveloppe" |

> **Do not add** `envelopes.empty` or `envelopes.emptyDesc` — they already exist and should be reused.

#### `insights` additions (only `insights.crossMission` exists — all below are NEW)

| Key | EN | FR |
|-----|----|----|
| `insights.loading` | "Analysis in progress…" | "Analyse en cours…" |
| `insights.noData` | "No data" | "Aucune donnée" |
| `insights.noDataDesc` | "Create missions to see your insights." | "Créez des missions pour voir vos insights." |
| `insights.title` | "Budget Insights" | "Insights Budget" |
| `insights.activeMissions` | "Active missions" | "Missions actives" |
| `insights.totalBudget` | "Total budget" | "Budget total" |
| `insights.budgetUsage` | "Budget usage" | "Utilisation budget" |
| `insights.recommendations` | "Recommendations" | "Recommandations" |
| `insights.byCategory` | "By category" | "Par catégorie" |
| `insights.allMissions` | "All missions" | "Toutes les missions" |

#### `loyalty` additions (namespace currently empty)

| Key | EN | FR |
|-----|----|----|
| `loyalty.title` | "My Loyalty Points" | "Mes Points Fidélité" |
| `loyalty.subtitle` | "Track your loyalty programmes" | "Suivi de vos programmes de fidélité" |

---

## Out of Scope

- `de.json` (German stub — falls back to EN, no native DE strings committed)
- Backend strings
- Pages not in the 4 identified above

---

## Success Criteria

1. `npm run lint` passes with zero `i18next/no-literal-string` errors after the fix pass
2. All 4 affected pages render correctly in both EN and FR locales (manual QA: switch language in Settings, verify text changes)
3. Any future PR that introduces a hardcoded JSX string will fail `npm run lint`

---

## Implementation Order

1. Install `eslint-plugin-i18next` as devDependency: `npm install -D eslint-plugin-i18next`
2. Update `eslint.config.js` with the config block from Part 1 above
3. Run `npm run lint` — confirm violations surface for the 4 pages
4. Add all new keys from the tables above to `en.json` and `fr.json`
5. Fix `LoyaltyPage.tsx` — add `useTranslation` import + hook, wrap 2 strings (smallest file, smoke test)
6. Fix `InsightsPage.tsx` — `useTranslation` already imported; wrap ~10 strings
7. Fix `CashbackPage.tsx` — add `useTranslation`, wrap ~18 strings including the interpolated `aria-label`
8. Fix `EnvelopesPage.tsx` — `useTranslation` already imported (do not duplicate); wrap ~16 strings, reuse existing `envelopes.empty`/`envelopes.emptyDesc`
9. Run `npm run lint` — expect zero violations
10. Run `npm run build` — confirm no TypeScript errors
