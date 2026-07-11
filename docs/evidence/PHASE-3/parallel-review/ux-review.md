# UX / Accessibility Review (cold-start)

Date: 2026-07-11T08:17:36Z
Reviewer: ux-reviewer (focused lens)

## Scope
- apps/frontend/app/* (routes)
- apps/frontend/components/* (UI primitives)
- styles/globals.css
- Playwright e2e screenshots from PHASE-3/screenshots/

## Findings

### 1. Contrast (color)


Concerns: with cyan/violet/gold/rose as primary accents on canvas #070A12, ensure AA contrast for body text (#E6EDF7). Spot-check: #E6EDF7 on #070A12 = ratio ~16:1 ✓.

### 2. Keyboard navigation


Concerns: most surfaces are static text or <Link>. No keyboard handlers. <Link> elements get default focus rings (browser native).
Verify focus rings are visible — DESIGN.md §1 forbids removing them.

### 3. ARIA labels

apps/frontend/components/kpi-tile.tsx:46:        aria-live="polite"
apps/frontend/components/kpi-tile.tsx:47:        aria-label={`${label} ${text}${unit ?? ""}`}

Coverage:
- KPITile has aria-live=\"polite\" + aria-label ✓
- TradeFeed / SocialFeed / WealthChart: no ARIA roles or live regions.
- Agent detail page vitals (DL): uses <dt>/<dd>, semantically correct.

### 4. Reduced motion

Concerns: DESIGN.md §6 mandates reduced-motion respect. globals.css does NOT implement prefers-reduced-motion. Framer Motion has its own reduce-motion handling but not explicitly wired here.

### 5. Color-only signal

Trade broadcast: text color + kind word.
6:const KIND_COLOR: Record<string, string> = {
24:          <span className={clsx("truncate", KIND_COLOR[it.kind])}>{it.text}</span>
Social feed: bracketed intent + content.
7:const INTENT_COLOR: Record<string, string> = {
60:            <span className={clsx("mr-2", INTENT_COLOR[p.intent_type])}>
Verdict: not color-only. Good.

### 6. Empty states

TradeFeed empty: \"等待第一笔交易...\"
SocialFeed empty: \"等待 Agent 们的第一条广播...\"
WealthChart empty: \"无 Agent 数据\"
All three are present and not loaders (good).

### 7. Error states

No error UI for fetch failures (e.g. backend down). KPI tile shows \"—\" / 0.
Concerns: silent failure mode. Add a top-level error banner when fetch fails repeatedly.

### 8. Typography

Display: Space Grotesk (per DESIGN.md)
Body: Inter
Mono: JetBrains Mono
globals.css defines --font-display/body/mono but no @font-face imports. Falls back to system-ui.
Concerns: fonts are declared but not loaded. Demo screenshot uses system fonts.
Low priority but DESIGN.md §3 specifies these fonts.

### 9. Mobile UX

grid-cols-2 on mobile (KPI), grid-cols-1 for body sections.
Job-type cards stack vertically.
Trade broadcast items wrap with truncate.
Verdict: works, dense but readable.

## Verdict
Dashboard is visually clean and operational.
Top 3 actionable items:
1. Add prefers-reduced-motion handling (DESIGN.md §6 compliance).
2. Add ARIA live regions on TradeFeed + SocialFeed.
3. Add a top-level error banner when fetch fails repeatedly.
