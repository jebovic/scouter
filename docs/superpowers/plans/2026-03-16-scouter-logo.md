# SCOUTER Logo Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the production-ready SCOUTER Scout Lens logo across static assets, a reusable React component, and nav integrations.

**Architecture:** Four independent layers — (1) static SVG assets (favicon, PWA icon, sprite), (2) `Logo` React component with three size variants and full accessibility, (3) NavRail integration replacing the hardcoded text/letter, (4) Topnav integration replacing the `nav.hq` text link.

**Tech Stack:** React 19 (`useId`), TypeScript, CSS Modules, Vitest + Testing Library, SVG

**Spec:** `docs/superpowers/specs/2026-03-16-scouter-logo-design.md`

---

## File Map

| File | Action | Notes |
|---|---|---|
| `frontend/public/favicon.svg` | Replace | Simplified mark — ring, dot, handle, no trend line |
| `frontend/public/icons/icon.svg` | Replace | Simplified mark on 512×512 dark rounded-rect canvas |
| `frontend/public/icons.svg` | Modify | Append two `<symbol>` blocks |
| `frontend/public/manifest.json` | Modify | `theme_color` `#00d4ff` → `#00e5ff` |
| `frontend/src/components/scouter/Logo.tsx` | Create | React component per spec API |
| `frontend/src/components/scouter/Logo.module.css` | Create | Gradient wordmark, size variants |
| `frontend/src/components/scouter/Logo.test.tsx` | Create | Unit tests: 8 cases |
| `frontend/src/components/scouter/index.ts` | Modify | Add `Logo` export |
| `frontend/src/components/scouter/NavRail.tsx` | Modify | Replace hardcoded SCOUTER text with `<Logo>` |
| `frontend/src/components/scouter/NavRail.module.css` | Modify | Remove `.logoText`, update `.logo` for flex alignment |
| `frontend/src/components/scouter/Topnav.tsx` | Modify | Replace text link with `<Logo size="md" />` |
| `frontend/src/components/scouter/Topnav.module.css` | Modify | Remove text-only styles from `.logo` |

---

## Chunk 1: Static Assets

**Files:**
- Replace: `frontend/public/favicon.svg`
- Replace: `frontend/public/icons/icon.svg`
- Modify: `frontend/public/icons.svg`
- Modify: `frontend/public/manifest.json`

No tests needed — these are static SVG/JSON files. Visual verification is via browser.

---

- [ ] **Step 1.1 — Replace favicon.svg**

  The current file is a complex lightning bolt SVG. Replace it entirely with the simplified Scout Lens mark (≤24px variant from the spec).

  Write `frontend/public/favicon.svg`:

  ```svg
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 80 80" fill="none"
       role="img" aria-label="SCOUTER">
    <defs>
      <linearGradient id="scouter-grad-sm" x1="8" y1="8" x2="72" y2="72" gradientUnits="userSpaceOnUse">
        <stop offset="0%" stop-color="#00e5ff"/>
        <stop offset="100%" stop-color="#a855f7"/>
      </linearGradient>
    </defs>
    <!-- Outer ring (thicker stroke for small render size) -->
    <circle cx="34" cy="34" r="22" stroke="url(#scouter-grad-sm)" stroke-width="5" fill="none"/>
    <!-- Handle -->
    <line x1="51" y1="51" x2="67" y2="67"
      stroke="url(#scouter-grad-sm)" stroke-width="7" stroke-linecap="round"/>
    <!-- Simplified center dot -->
    <circle cx="34" cy="34" r="6" fill="url(#scouter-grad-sm)"/>
  </svg>
  ```

- [ ] **Step 1.2 — Replace icon.svg**

  The current file is a `100×100` viewBox with a telescope emoji. Replace with the simplified mark centred inside a `512×512` rounded-rect canvas (`rx="102"`, ~20% of 512), matching the PWA app icon spec.

  Write `frontend/public/icons/icon.svg`:

  ```svg
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" fill="none"
       role="img" aria-label="SCOUTER">
    <defs>
      <linearGradient id="scouter-icon-grad" x1="100" y1="100" x2="412" y2="412" gradientUnits="userSpaceOnUse">
        <stop offset="0%" stop-color="#00e5ff"/>
        <stop offset="100%" stop-color="#a855f7"/>
      </linearGradient>
    </defs>
    <!-- Dark rounded-rect background -->
    <rect width="512" height="512" rx="102" fill="#0a0e1a"/>
    <!--
      Simplified Scout Lens mark scaled to ~300×300, centred at (256,256).
      Original viewBox: 0 0 80 80. Scale factor: 300/80 = 3.75.
      Translate: (256 - 34*3.75) = 256 - 127.5 = 128.5 for cx,
                 so origin shift = (256 - 40*3.75) = 256 - 150 = 106
      Applied via transform="translate(106,106) scale(3.75)"
    -->
    <g transform="translate(106,106) scale(3.75)">
      <!-- Outer ring -->
      <circle cx="34" cy="34" r="22" stroke="url(#scouter-icon-grad)" stroke-width="5" fill="none"/>
      <!-- Handle -->
      <line x1="51" y1="51" x2="67" y2="67"
        stroke="url(#scouter-icon-grad)" stroke-width="7" stroke-linecap="round"/>
      <!-- Center dot -->
      <circle cx="34" cy="34" r="6" fill="url(#scouter-icon-grad)"/>
    </g>
  </svg>
  ```

- [ ] **Step 1.3 — Add sprite symbols to icons.svg**

  The existing `frontend/public/icons.svg` starts with `<svg xmlns="http://www.w3.org/2000/svg">` and contains existing `<symbol>` entries. Append two new symbols just before the closing `</svg>` tag.

  **Why `currentColor`:** `<use>` referencing an external sprite file cannot inherit `linearGradient` definitions from the parent document. These sprite symbols are for non-React usage (HTML, email). The `<Logo>` React component inlines SVG directly and does NOT use `<use>`.

  Add before the closing `</svg>`:

  ```svg
  <!-- Full detail Scout Lens mark (32px+) — uses SVG clipPath -->
  <symbol id="scouter-icon" viewBox="0 0 80 80">
    <defs>
      <clipPath id="scouter-icon-clip">
        <circle cx="34" cy="34" r="21"/>
      </clipPath>
    </defs>
    <circle cx="34" cy="34" r="22" stroke="currentColor" stroke-width="3" fill="none"/>
    <g clip-path="url(#scouter-icon-clip)">
      <polyline
        points="16,42 22,38 27,41 32,33 37,30 42,24 48,22"
        stroke="currentColor" stroke-width="2.5"
        stroke-linecap="round" stroke-linejoin="round"
        fill="none" opacity="0.85"
      />
      <circle cx="48" cy="22" r="2.5" fill="currentColor"/>
    </g>
    <line x1="51" y1="51" x2="67" y2="67" stroke="currentColor" stroke-width="4" stroke-linecap="round"/>
  </symbol>

  <!-- Simplified Scout Lens mark (≤24px) -->
  <symbol id="scouter-icon-sm" viewBox="0 0 80 80">
    <circle cx="34" cy="34" r="22" stroke="currentColor" stroke-width="5" fill="none"/>
    <line x1="51" y1="51" x2="67" y2="67" stroke="currentColor" stroke-width="7" stroke-linecap="round"/>
    <circle cx="34" cy="34" r="6" fill="currentColor"/>
  </symbol>
  ```

- [ ] **Step 1.4 — Update manifest.json theme_color**

  Current: `"theme_color": "#00d4ff"`
  Change to: `"theme_color": "#00e5ff"`

  This aligns with the actual design-system cyan token (`--cyan: #00e5ff`).

- [ ] **Step 1.5 — Commit static assets**

  ```bash
  git add frontend/public/favicon.svg frontend/public/icons/icon.svg \
          frontend/public/icons.svg frontend/public/manifest.json
  git commit -m "feat(logo): Scout Lens favicon, PWA icon, sprite symbols, manifest theme_color"
  ```

---

## Chunk 2: Logo React Component

**Files:**
- Create: `frontend/src/components/scouter/Logo.tsx`
- Create: `frontend/src/components/scouter/Logo.module.css`
- Create: `frontend/src/components/scouter/Logo.test.tsx`
- Modify: `frontend/src/components/scouter/index.ts` (add one export line)

---

- [ ] **Step 2.1 — Write the failing tests**

  Create `frontend/src/components/scouter/Logo.test.tsx`:

  ```tsx
  import { describe, it, expect } from 'vitest'
  import { render, screen } from '@testing-library/react'
  import { Logo } from './Logo'

  describe('Logo', () => {
    it('renders wordmark text by default', () => {
      render(<Logo />)
      expect(screen.getByText('SCOUTER')).toBeInTheDocument()
    })

    it('hides icon from accessibility tree when wordmark is visible', () => {
      const { container } = render(<Logo />)
      const svg = container.querySelector('svg')
      expect(svg).toHaveAttribute('aria-hidden', 'true')
    })

    it('renders tagline at size="lg" when showTagline=true', () => {
      render(<Logo size="lg" showTagline />)
      expect(screen.getByText('spending intelligence')).toBeInTheDocument()
    })

    it('does not render tagline at size="md" even when showTagline=true', () => {
      render(<Logo size="md" showTagline />)
      expect(screen.queryByText('spending intelligence')).not.toBeInTheDocument()
    })

    it('does not render tagline at size="sm" even when showTagline=true', () => {
      render(<Logo size="sm" showTagline />)
      expect(screen.queryByText('spending intelligence')).not.toBeInTheDocument()
    })

    it('iconOnly: no wordmark, SVG has role="img" and default aria-label', () => {
      const { container } = render(<Logo iconOnly />)
      expect(screen.queryByText('SCOUTER')).not.toBeInTheDocument()
      const svg = container.querySelector('svg')
      expect(svg).toHaveAttribute('role', 'img')
      expect(svg).toHaveAttribute('aria-label', 'SCOUTER')
    })

    it('iconOnly: custom aria-label overrides default', () => {
      const { container } = render(<Logo iconOnly aria-label="Scout Logo" />)
      const svg = container.querySelector('svg')
      expect(svg).toHaveAttribute('aria-label', 'Scout Logo')
    })

    it('multiple instances produce unique gradient ids', () => {
      const { container } = render(
        <>
          <Logo />
          <Logo />
        </>
      )
      const gradients = container.querySelectorAll('linearGradient')
      const ids = Array.from(gradients).map(g => g.id)
      const unique = new Set(ids)
      expect(unique.size).toBe(ids.length)
    })
  })
  ```

- [ ] **Step 2.2 — Run tests to verify they fail**

  ```bash
  cd frontend && npx vitest run src/components/scouter/Logo.test.tsx
  ```

  Expected: FAIL — `Logo` module not found.

- [ ] **Step 2.3 — Create Logo.module.css**

  Create `frontend/src/components/scouter/Logo.module.css`:

  ```css
  /* ── Root container ─────────────────────────────────── */
  .root {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    flex-shrink: 0;
    text-decoration: none;
    line-height: 1;
  }

  /* ── Text column (wordmark + optional tagline) ──────── */
  .text {
    display: flex;
    flex-direction: column;
  }

  /* ── Wordmark ───────────────────────────────────────── */
  .wordmark {
    font-family: 'Chakra Petch', sans-serif;
    font-weight: 700;
    letter-spacing: 0.14em;
    background: linear-gradient(90deg, var(--logo-start, #00e5ff), var(--logo-end, #a855f7));
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    line-height: 1.1;
    white-space: nowrap;
  }

  /* ── Tagline ────────────────────────────────────────── */
  .tagline {
    font-family: 'Chakra Petch', sans-serif;
    font-size: 9px;
    letter-spacing: 0.28em;
    color: var(--text-dim, #727f9e);
    text-transform: uppercase;
    margin-top: 2px;
    white-space: nowrap;
  }

  /* ── Size variants ──────────────────────────────────── */
  .sm .wordmark { font-size: 15px; }
  .md .wordmark { font-size: 20px; }
  .lg .wordmark { font-size: 28px; }
  ```

- [ ] **Step 2.4 — Create Logo.tsx**

  Create `frontend/src/components/scouter/Logo.tsx`:

  ```tsx
  import { useId } from 'react'
  import styles from './Logo.module.css'

  export interface LogoProps {
    size?: 'sm' | 'md' | 'lg'
    showTagline?: boolean
    iconOnly?: boolean
    variant?: 'dark' | 'light'
    gradientId?: string
    className?: string
    'aria-label'?: string
  }

  const ICON_SIZE: Record<NonNullable<LogoProps['size']>, number> = {
    sm: 28,
    md: 40,
    lg: 52,
  }

  export function Logo({
    size = 'md',
    showTagline = false,
    iconOnly = false,
    variant = 'dark',
    gradientId,
    className,
    'aria-label': ariaLabel = 'SCOUTER',
  }: LogoProps) {
    const autoId = useId()
    const prefix = gradientId ?? autoId.replace(/:/g, 'id')
    const gradId = `${prefix}-grad`
    const clipId = `${prefix}-lens`
    const px = ICON_SIZE[size]

    const c1 = variant === 'light' ? '#0099bb' : '#00e5ff'
    const c2 = variant === 'light' ? '#7c3aed' : '#a855f7'

    const svgSharedProps = iconOnly
      ? { role: 'img' as const, 'aria-label': ariaLabel }
      : { 'aria-hidden': true as const }

    const icon =
      size === 'sm' ? (
        // Simplified mark — no trend line (too fine at ≤28px)
        <svg
          width={px}
          height={px}
          viewBox="0 0 80 80"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          {...svgSharedProps}
        >
          <defs>
            <linearGradient id={gradId} x1="8" y1="8" x2="72" y2="72" gradientUnits="userSpaceOnUse">
              <stop offset="0%" stopColor={c1} />
              <stop offset="100%" stopColor={c2} />
            </linearGradient>
          </defs>
          <circle cx="34" cy="34" r="22" stroke={`url(#${gradId})`} strokeWidth="5" fill="none" />
          <line x1="51" y1="51" x2="67" y2="67" stroke={`url(#${gradId})`} strokeWidth="7" strokeLinecap="round" />
          <circle cx="34" cy="34" r="6" fill={`url(#${gradId})`} />
        </svg>
      ) : (
        // Full detail mark with price trend line inside lens
        <svg
          width={px}
          height={px}
          viewBox="0 0 80 80"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          {...svgSharedProps}
        >
          <defs>
            <linearGradient id={gradId} x1="8" y1="8" x2="72" y2="72" gradientUnits="userSpaceOnUse">
              <stop offset="0%" stopColor={c1} />
              <stop offset="100%" stopColor={c2} />
            </linearGradient>
            <clipPath id={clipId}>
              <circle cx="34" cy="34" r="21" />
            </clipPath>
          </defs>
          {/* Outer lens ring */}
          <circle cx="34" cy="34" r="22" stroke={`url(#${gradId})`} strokeWidth="3" fill="none" />
          {/* Price trend line clipped to lens interior */}
          <g clipPath={`url(#${clipId})`}>
            <polyline
              points="16,42 22,38 27,41 32,33 37,30 42,24 48,22"
              stroke={c1}
              strokeWidth="2.5"
              strokeLinecap="round"
              strokeLinejoin="round"
              fill="none"
              opacity={0.85}
            />
            {/* Trend tip dot */}
            <circle cx="48" cy="22" r="2.5" fill={c1} />
          </g>
          {/* Handle */}
          <line x1="51" y1="51" x2="67" y2="67" stroke={`url(#${gradId})`} strokeWidth="4" strokeLinecap="round" />
        </svg>
      )

    const gradientVars = {
      '--logo-start': c1,
      '--logo-end': c2,
    } as React.CSSProperties

    if (iconOnly) {
      return (
        <span
          className={[styles.root, styles[size], className].filter(Boolean).join(' ')}
          style={gradientVars}
        >
          {icon}
        </span>
      )
    }

    return (
      <span
        className={[styles.root, styles[size], className].filter(Boolean).join(' ')}
        style={gradientVars}
      >
        {icon}
        <span className={styles.text}>
          <span className={styles.wordmark}>SCOUTER</span>
          {size === 'lg' && showTagline && (
            <span className={styles.tagline}>spending intelligence</span>
          )}
        </span>
      </span>
    )
  }
  ```

- [ ] **Step 2.5 — Run tests to verify they pass**

  ```bash
  cd frontend && npx vitest run src/components/scouter/Logo.test.tsx
  ```

  Expected: all 8 tests PASS.

- [ ] **Step 2.6 — Add Logo export to index.ts**

  In `frontend/src/components/scouter/index.ts`, add one line after the last export:

  ```ts
  export { Logo } from './Logo'
  ```

  (Append after line 40: `export { NextActionNudge } from './NextActionNudge'`)

- [ ] **Step 2.7 — Run typecheck**

  ```bash
  cd frontend && npm run typecheck
  ```

  Expected: no errors.

- [ ] **Step 2.8 — Commit Logo component**

  ```bash
  git add frontend/src/components/scouter/Logo.tsx \
          frontend/src/components/scouter/Logo.module.css \
          frontend/src/components/scouter/Logo.test.tsx \
          frontend/src/components/scouter/index.ts
  git commit -m "feat(logo): Logo React component — Scout Lens, 3 sizes, gradient wordmark, accessibility"
  ```

---

## Chunk 3: NavRail Integration

**Files:**
- Modify: `frontend/src/components/scouter/NavRail.tsx` (lines 111–116)
- Modify: `frontend/src/components/scouter/NavRail.module.css` (lines 29–49)

**Context:** NavRail currently renders `{collapsed ? 'S' : 'SCOUTER'}` as a plain text span inside a `<Link>`. We replace this with `<Logo size="sm" />` (expanded) and `<Logo size="sm" iconOnly />` (collapsed).

---

- [ ] **Step 3.1 — Update NavRail.tsx logo section**

  In `frontend/src/components/scouter/NavRail.tsx`:

  1. Add import at top (after existing imports):

  ```tsx
  import { Logo } from './Logo'
  ```

  2. Replace lines 111–116 (the `{/* Logo */}` block):

  **Before:**
  ```tsx
  {/* Logo */}
  <Link to="/" className={styles.logo} title={t('nav.logoTitle')}>
    <span className={styles.logoText}>
      {collapsed ? 'S' : 'SCOUTER'}
    </span>
  </Link>
  ```

  **After:**
  ```tsx
  {/* Logo */}
  <Link to="/" className={styles.logo} title={t('nav.logoTitle')}>
    {collapsed
      ? <Logo size="sm" iconOnly aria-label={t('nav.logoTitle')} />
      : <Logo size="sm" />
    }
  </Link>
  ```

- [ ] **Step 3.2 — Update NavRail.module.css logo styles**

  In `frontend/src/components/scouter/NavRail.module.css`, the `.logo` block (lines 30–38) and `.logoText` block (lines 41–49) need updating. The `.logoText` class is no longer needed. The `.logo` block needs a `gap` for the logo component.

  **Before:**
  ```css
  /* ── Logo ─────────────────────────────────────────────── */
  .logo {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 52px;
    padding: 0 12px;
    border-bottom: 1px solid var(--border);
    text-decoration: none;
    flex-shrink: 0;
  }

  .logoText {
    font-family: var(--font-display);
    font-size: 1rem;
    font-weight: 700;
    letter-spacing: 0.12em;
    color: var(--cyan);
    white-space: nowrap;
    overflow: hidden;
  }
  ```

  **After:**
  ```css
  /* ── Logo ─────────────────────────────────────────────── */
  .logo {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 52px;
    padding: 0 12px;
    border-bottom: 1px solid var(--border);
    text-decoration: none;
    flex-shrink: 0;
    overflow: hidden;
  }
  ```

  (Remove the entire `.logoText` block — the Logo component owns its own text styling.)

- [ ] **Step 3.3 — Run full frontend test suite**

  ```bash
  cd frontend && npm run test
  ```

  Expected: all tests pass (no NavRail test file exists, but existing tests must not regress).

- [ ] **Step 3.4 — Run typecheck**

  ```bash
  cd frontend && npm run typecheck
  ```

  Expected: no errors.

- [ ] **Step 3.5 — Commit NavRail update**

  ```bash
  git add frontend/src/components/scouter/NavRail.tsx \
          frontend/src/components/scouter/NavRail.module.css
  git commit -m "feat(logo): integrate Logo component into NavRail"
  ```

---

## Chunk 4: Topnav Integration

**Files:**
- Modify: `frontend/src/components/scouter/Topnav.tsx` (lines 34–37)
- Modify: `frontend/src/components/scouter/Topnav.module.css` (lines 41–50)

**Context:** The Topnav currently renders `{t('nav.hq')}` (the text "HQ") inside a `<Link className={styles.logo}>`. We replace this with `<Logo size="md" />`. The translation key `nav.hq` is no longer used for the logo — it may still be used elsewhere (mission sub-nav breadcrumb area) so do NOT delete it from i18n files; just stop using it here.

---

- [ ] **Step 4.1 — Update Topnav.tsx logo section**

  In `frontend/src/components/scouter/Topnav.tsx`:

  1. Add import (after existing imports):

  ```tsx
  import { Logo } from './Logo'
  ```

  2. Replace lines 34–37 (the `{/* Logo */}` block):

  **Before:**
  ```tsx
  {/* Logo */}
  <Link to="/" className={styles.logo}>
    {t('nav.hq')}
  </Link>
  ```

  **After:**
  ```tsx
  {/* Logo */}
  <Link to="/" className={styles.logo} aria-label="SCOUTER — Home">
    <Logo size="md" />
  </Link>
  ```

- [ ] **Step 4.2 — Update Topnav.module.css logo styles**

  The current `.logo` class in `Topnav.module.css` (lines 41–50) has text-specific styles (`font-family`, `color`, `letter-spacing`, `text-shadow`). Since `<Logo>` owns all its own styling, we strip those out and keep only the layout properties needed to seat the component in the nav bar.

  **Before:**
  ```css
  .logo {
    font-family: var(--font-display);
    color: var(--cyan);
    letter-spacing: 0.15em;
    font-size: 0.95rem;
    font-weight: 700;
    flex-shrink: 0;
    text-shadow: 0 0 12px var(--cyan-dim);
    text-decoration: none;
  }
  ```

  **After:**
  ```css
  .logo {
    flex-shrink: 0;
    text-decoration: none;
    display: flex;
    align-items: center;
  }
  ```

- [ ] **Step 4.3 — Run full frontend test suite**

  ```bash
  cd frontend && npm run test
  ```

  Expected: all tests pass.

- [ ] **Step 4.4 — Run typecheck**

  ```bash
  cd frontend && npm run typecheck
  ```

  Expected: no errors.

- [ ] **Step 4.5 — Run build**

  ```bash
  cd frontend && npm run build
  ```

  Expected: clean build, no warnings about unused imports.

- [ ] **Step 4.6 — Commit Topnav update**

  ```bash
  git add frontend/src/components/scouter/Topnav.tsx \
          frontend/src/components/scouter/Topnav.module.css
  git commit -m "feat(logo): integrate Logo component into Topnav"
  ```

---

## Final Verification

- [ ] **Step 5.1 — Run full test suite + typecheck + build**

  ```bash
  cd frontend && npm run test && npm run typecheck && npm run build
  ```

  Expected: all pass cleanly.

- [ ] **Step 5.2 — Visual check in browser**

  Start the dev server (`npm run dev` or `make dev`) and verify:
  - Favicon in browser tab shows the Scout Lens mark (ring + dot + handle, cyan→purple gradient)
  - NavRail expanded: `<Logo size="sm" />` with icon + "SCOUTER" wordmark
  - NavRail collapsed: `<Logo size="sm" iconOnly />` showing icon only
  - Topnav: `<Logo size="md" />` — slightly larger icon + wordmark
  - No duplicate gradient IDs visible in browser DevTools (inspect SVG elements)
