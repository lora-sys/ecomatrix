# DESIGN.md — EcoMatrix Visual & Interaction System

> Source of truth for brand, tokens, components, and motion.
> Implemented in `apps/frontend` (Next.js + Tailwind + Aceternity UI + Framer Motion).

## 1. Brand Pillars

1. **God's-Eye Observatory** — the operator is an omnipotent observer, never a participant.
2. **Economic Pressure** — every visual must imply consequence: wealth gradients, decay, heat.
3. **Latent Cyberpunk** — neon accents on a deep neutral canvas; never a 90s arcade cliche.
4. **Density over decoration** — packed, scannable data; quiet chrome; bold signal only where it matters.

Forbidden aesthetics: glassmorphism blobs, gradient orbs, decorative cards around hero text, rounded "playful" UI for an operational tool.

## 2. Color Tokens

Defined as Tailwind theme tokens (`apps/frontend/tailwind.config.ts`).

| Token              | Hex       | Usage                                            |
| ------------------ | --------- | ------------------------------------------------ |
| `bg-canvas`        | `#070A12` | App background; the "void"                        |
| `bg-panel`         | `#0E1422` | Surface (cards, drawers)                         |
| `bg-panel-2`       | `#141B2D` | Elevated surface                                 |
| `border-hairline`  | `#1F2A44` | 1px dividers                                     |
| `text-primary`     | `#E6EDF7` | Primary text                                     |
| `text-muted`       | `#8A95B2` | Secondary, labels                                |
| `text-dim`         | `#5A6685` | Tertiary, captions                               |
| `accent-cyan`      | `#22D3EE` | Live data, primary signal                        |
| `accent-violet`    | `#7C5CFF` | Secondary signal, hover                          |
| `accent-gold`      | `#F5C044` | Wealth, "GOLD"                                   |
| `accent-rose`      | `#F43F5E` | Alerts, failed trades                            |
| `accent-emerald`   | `#34D399` | Positive delta, settled tx                       |
| `grad-wealth`      | `cyan→gold→rose` | Wealth distribution gradient (poor→rich→over-leveraged) |

No pure black, no pure white. No purple-blue gradient washes.

## 3. Typography

- Display: **Space Grotesk** (weights 500/600/700) — used for KPI numerals and section titles only.
- Body / UI: **Inter** (400/500/600).
- Mono: **JetBrains Mono** — agent IDs, tx hashes, code, JSON pretty-print.
- Never use `clamp()` to scale font to viewport. Use discrete size tokens.
- Letter spacing: 0. Tracking must never be negative.

## 4. Spacing & Layout

- 4 px base. Tokens: `1, 2, 3, 4, 6, 8, 12, 16, 24` (Tailwind default).
- Container max width: 1440 px. Grids use 12 cols at ≥1024 px, 6 cols at ≥640 px, 4 cols below.
- Density rule: dashboard pages must show ≥3 distinct data signals above the fold on a 1440×900 desktop.

## 5. Core Components (Aceternity-mapped)

| Surface                | Aceternity primitive   | Notes                                                |
| ---------------------- | ---------------------- | ---------------------------------------------------- |
| KPI tile               | `Glowing Cards`        | Single dominant metric, neon ring, value + delta.    |
| Social timeline        | `Tracing Beam`         | Vertical neon thread; node color = agent job.        |
| Agent detail           | `3D Card Effect`       | Tilt on hover (perspective ≤ 8°), gold ring on focus.|
| Trade feed             | `Background Gradient` + monospace rows | No card chrome around the feed itself. |
| Modal / drawer         | Plain panel + hairline | No floating "glass" cards.                           |
| Empty / error          | Plain text + icon      | Never an illustration.                               |

## 6. Motion

- **Framer Motion** is the only motion library.
- Defaults: 180 ms ease-out for state transitions; 320 ms cubic-bezier(0.2, 0.8, 0.2, 1) for layout shifts.
- KPI numerals: `useMotionValue` + `useTransform` to tick-count, never flash-cut.
- WebSocket updates must use a **value damping** layer (300 ms) to avoid flicker storms.
- Reduced motion (`prefers-reduced-motion: reduce`) must disable tilt and tracing-beam animation.

## 7. Iconography

- **lucide-react** for everything. No emoji in UI. No hand-drawn SVG icons.
- Job glyphs (miner / merchant / hacker / mediator) are color-coded chips, not illustrations.

## 8. Data Visualization

- Charts: `recharts` (line, area, scatter). Wealth distribution must use the `grad-wealth` color scale.
- The trading broadcast is a **scrolling monospace log**, not a chart.
- No 3D pie charts. No donut charts for live data.

## 9. Accessibility

- WCAG 2.1 AA contrast minimum on `bg-canvas`. Cyan/violet accents must not be the sole signal.
- Keyboard: every interactive surface is reachable; focus ring uses `accent-cyan` at 2 px.
- Screen reader: KPI deltas expose `aria-live="polite"` and `aria-label` with absolute change.

## 10. Internationalization

- Primary: `zh-CN` (matches PRD demo locale).
- Strings live in `apps/frontend/messages/{zh-CN,en}.json`. No hardcoded user-visible strings.
- Currency label: `金币` (zh-CN) / `GOLD` (en).

## 11. Forbidden Patterns

- ❌ Card-on-card layouts (don't nest a card inside another card).
- ❌ Hero gradient backgrounds with floating orbs.
- ❌ Marketing-style "feature spotlight" pages — this is an operational tool.
- ❌ Scaling font size with viewport width.
- ❌ Single-hue palettes (all-purple, all-blue).
- ❌ Visible text describing how the app works — the UI must be self-evident.
