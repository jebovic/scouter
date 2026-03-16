# SCOUTER Logo Design Spec

**Date:** 2026-03-16
**Status:** Approved
**Author:** Brainstorming session

---

## Overview

Design a production-ready logo for SCOUTER Universal — a personal spending intelligence tool. The logo must be distinctive, scalable from 16px to display size, and consistent with the existing dark-mode design system.

---

## Design Decisions

| Dimension | Decision | Rationale |
|---|---|---|
| Personality | Smart Companion | Friendly, approachable, clever — like Notion/Arc Browser. Dark-mode sophisticated without being cold. |
| Icon concept | Scout Lens | Magnifying glass / crosshair hybrid — "searching with intelligence." |
| Differentiator | Price trend line inside lens | Eliminates confusion with generic search icons. Communicates "we analyse data, not just find things." |
| Color | Cyan → Purple gradient | `#00e5ff` → `#a855f7`. Echoes existing favicon palette. Dynamic and modern. |
| Layout | Horizontal lock-up | Icon left, wordmark right. Standard for nav bars and app headers. |
| Wordmark | `SCOUTER` in Chakra Petch 700 | Uses `--font-body` token (`'Chakra Petch', sans-serif`). Letter-spacing: 0.14em. |
| Tagline | `spending intelligence` | Shown at `lg` size only. Chakra Petch, 9px, letter-spacing 0.28em, `--text-dim` color. |

---

## Icon Specification

### SVG Mark — Full Detail (32px+)

Used in horizontal lock-up at `md` and `lg` sizes.

```svg
<svg viewBox="0 0 80 80" fill="none" xmlns="http://www.w3.org/2000/svg"
     role="img" aria-label="SCOUTER">
  <defs>
    <linearGradient id="scouter-grad" x1="8" y1="8" x2="72" y2="72" gradientUnits="userSpaceOnUse">
      <stop offset="0%" stop-color="#00e5ff"/>
      <stop offset="100%" stop-color="#a855f7"/>
    </linearGradient>
    <clipPath id="scouter-lens">
      <circle cx="34" cy="34" r="21"/>
    </clipPath>
  </defs>

  <!-- Outer lens ring -->
  <circle cx="34" cy="34" r="22" stroke="url(#scouter-grad)" stroke-width="3" fill="none"/>

  <!-- Price trend line (clipped to lens interior) -->
  <g clip-path="url(#scouter-lens)">
    <polyline
      points="16,42 22,38 27,41 32,33 37,30 42,24 48,22"
      stroke="#00e5ff"
      stroke-width="2.5"
      stroke-linecap="round"
      stroke-linejoin="round"
      fill="none"
      opacity="0.85"
    />
    <!-- Trend tip dot -->
    <circle cx="48" cy="22" r="2.5" fill="#00e5ff"/>
  </g>

  <!-- Handle -->
  <line x1="51" y1="51" x2="67" y2="67"
    stroke="url(#scouter-grad)" stroke-width="4" stroke-linecap="round"/>
</svg>
```

### SVG Mark — Simplified (≤24px: favicon, app icon, sidebar collapsed)

The trend line is too fine to render cleanly below 32px. Use this variant for favicon, the PWA app icon, and the NavRail collapsed state.

```svg
<svg viewBox="0 0 80 80" fill="none" xmlns="http://www.w3.org/2000/svg"
     role="img" aria-label="SCOUTER">
  <defs>
    <linearGradient id="scouter-grad-sm" x1="8" y1="8" x2="72" y2="72" gradientUnits="userSpaceOnUse">
      <stop offset="0%" stop-color="#00e5ff"/>
      <stop offset="100%" stop-color="#a855f7"/>
    </linearGradient>
  </defs>
  <!-- Outer ring (thicker stroke compensates for small render size) -->
  <circle cx="34" cy="34" r="22" stroke="url(#scouter-grad-sm)" stroke-width="5" fill="none"/>
  <!-- Handle -->
  <line x1="51" y1="51" x2="67" y2="67"
    stroke="url(#scouter-grad-sm)" stroke-width="7" stroke-linecap="round"/>
  <!-- Simplified center dot -->
  <circle cx="34" cy="34" r="6" fill="url(#scouter-grad-sm)"/>
</svg>
```

---

## Wordmark

- **Font:** explicit `font-family: 'Chakra Petch', sans-serif` (same as `--font-body` token), weight 700
- **Color:** CSS gradient `linear-gradient(90deg, #00e5ff, #a855f7)` via `-webkit-background-clip: text` / `background-clip: text`
- **Letter-spacing:** 0.14em
- **Case:** ALL CAPS

---

## Color Tokens

| Context | Gradient start | Gradient end | Trigger |
|---|---|---|---|
| Dark background (primary) | `#00e5ff` | `#a855f7` | Default |
| Light background | `#0099bb` | `#7c3aed` | `variant="light"` prop on `<Logo>`, or `@media (prefers-color-scheme: light)` |

The light variant shifts the gradient slightly darker so the stroke-based icon meets WCAG AA contrast on white (`#ffffff`) backgrounds.

---

## Size Variants

| Size | Icon detail | Wordmark | Tagline |
|---|---|---|---|
| `sm` — 28px icon | Simplified (no trend line) | Yes, 15px | No |
| `md` — 40px icon | Full detail | Yes, 20px | No |
| `lg` — 52px icon | Full detail | Yes, 28px | Yes, 9px |

---

## React Component API

```tsx
interface LogoProps {
  size?: 'sm' | 'md' | 'lg';        // default: 'md'
  showTagline?: boolean;             // only renders at size='lg'; silently ignored otherwise
  iconOnly?: boolean;                // render icon mark only, no wordmark (e.g. NavRail collapsed)
  variant?: 'dark' | 'light';       // default: 'dark'. Explicit prop takes precedence over prefers-color-scheme.
  gradientId?: string;               // override gradient id prefix for uniqueness (default: React useId())
  className?: string;
  'aria-label'?: string;             // default: 'SCOUTER'
}
```

**Gradient ID uniqueness:** The component uses React 19's `useId()` to generate a unique prefix for all gradient/clipPath `id` attributes, preventing conflicts when multiple `<Logo>` instances appear on the same page.

**`showTagline` behaviour:** When `size !== 'lg'`, the `showTagline` prop is silently ignored — the tagline is never rendered regardless of prop value. No prop type error is thrown.

**`variant` vs `prefers-color-scheme`:** The `variant` prop takes explicit precedence. Only when `variant` is omitted (or `undefined`) does the component fall back to a `prefers-color-scheme: light` media query check. Default is `'dark'` (no media query fallback unless explicitly opted in).

---

## Accessibility

- When `iconOnly={false}` (default — wordmark visible): icon SVG gets `aria-hidden="true"`; the wordmark `<span>` carries the accessible text. The wrapping element gets no extra role.
- When `iconOnly={true}` (no wordmark): icon SVG keeps `role="img"` and `aria-label` (defaulting to `"SCOUTER"`, overridable via the `aria-label` prop). This covers the NavRail collapsed state.
- Favicon and `icon.svg` are static SVG files — they carry `role="img"` and `aria-label="SCOUTER"` directly in the markup.

---

## Files to Create / Update

| File | Action | Notes |
|---|---|---|
| `frontend/public/favicon.svg` | **Replace** | Simplified mark (≤24px version) — favicon renders at 16px in browser tabs |
| `frontend/public/icons/icon.svg` | **Replace** | Simplified mark centred in `viewBox="0 0 512 512"` rounded rect (`rx="102"`, ~20% of width), `fill="#0a0e1a"` dark background. Icon mark scaled to ~300×300 and centred at (256,256). |
| `frontend/public/icons.svg` | **Update** | Add two `<symbol>` entries — see SVG sprite section below |
| `frontend/public/manifest.json` | **Update** | Update `theme_color` from `#00d4ff` → `#00e5ff` to match new gradient start |
| `frontend/src/components/scouter/Logo.tsx` | **Create** | React component per API above |
| `frontend/src/components/scouter/Logo.module.css` | **Create** | Gradient wordmark styles, size variants |
| `frontend/src/components/scouter/Logo.test.tsx` | **Create** | Unit tests: renders at each size, tagline visibility, aria-label, light variant, iconOnly mode |
| `frontend/src/components/scouter/index.ts` | **Update** | Add `export { Logo } from './Logo'` |
| NavRail (`frontend/src/components/scouter/NavRail.tsx`) | **Update** | Replace hardcoded `"SCOUTER"` text with `<Logo size="sm" />`. Replace collapsed `'S'` fallback with `<Logo size="sm" iconOnly />` |
| Topnav / GlobalLayout nav bar | **Update** | Replace placeholder with `<Logo size="md" />` |

---

## SVG Sprite Symbols (`icons.svg`)

Add these two `<symbol>` blocks inside the root `<svg>` of `frontend/public/icons.svg`.

**Important:** SVG `<use>` referencing an external sprite cannot inherit `linearGradient` from the document. Sprite symbols use `currentColor` throughout — this makes them suitable for monochrome/single-color contexts. The `<Logo>` React component does **not** use `<use>`; it inlines the SVG directly with `useId()`-prefixed gradient IDs. The sprite symbols exist for any future non-React usage (e.g., plain HTML pages, email templates).

```svg
<!-- Full detail mark (32px+) — uses SVG clipPath, not CSS clip-path -->
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

<!-- Simplified mark (≤24px) -->
<symbol id="scouter-icon-sm" viewBox="0 0 80 80">
  <circle cx="34" cy="34" r="22" stroke="currentColor" stroke-width="5" fill="none"/>
  <line x1="51" y1="51" x2="67" y2="67" stroke="currentColor" stroke-width="7" stroke-linecap="round"/>
  <circle cx="34" cy="34" r="6" fill="currentColor"/>
</symbol>
```

---

## Implementation Notes

- **React 19 confirmed** — `useId()` is available and preferred over manual ID management.
- The favicon file is `favicon.svg` (already an SVG, not `.ico`) — replace in-place.
- The `icon.svg` PWA app icon uses the simplified mark to ensure legibility at 192×192 and 512×512 manifest icon sizes.
- `manifest.json` `theme_color` update from `#00d4ff` → `#00e5ff` is a minor correction to align with the actual design-system cyan token.
- No `Oxanium` font is used in the logo — the wordmark explicitly uses `'Chakra Petch', sans-serif` matching the `--font-body` token. `--font-display` (Oxanium) is not used here.
