# Frontend Architecture (`apps/frontend`)

## Stack
- Next.js 15 (App Router), React 19, TypeScript 5 (strict), Tailwind CSS, Aceternity UI, Framer Motion, lucide-react, recharts, zustand, zod, Vitest, Playwright.

## Routes (Phase 3)
| Route             | Purpose                                                |
| ----------------- | ------------------------------------------------------ |
| `/`               | Dashboard: KPI tiles, wealth chart, trade broadcast.   |
| `/agents/[id]`    | Agent detail: vitals, CoT trace, recent trades.        |
| `/api/*`          | BFF routes (proxy to Go, never expose admin token).    |

## Data Flow
- Initial page load: RSC fetches via `lib/api/*.ts` (typed, uses `process.env.BACKEND_URL`).
- Live updates: a client-side WS hook (`hooks/use-economatrix-stream.ts`) subscribes once per page; per-component slices via `zustand`.
- Value damping: KPI tiles animate over 300 ms; never flash-cut to a new value.

## Components
- `components/kpi-tile.tsx`
- `components/wealth-chart.tsx`
- `components/trade-broadcast.tsx`
- `components/agent-card.tsx`
- `components/cot-trace.tsx`
- `components/tracing-beam.tsx` (wraps Aceternity)
- `components/glowing-card.tsx` (wraps Aceternity)

## i18n
`next-intl` with `messages/zh-CN.json` default. Tokens live in `DESIGN.md`.

## Tests
- Vitest: hooks, store reducers, schema validators.
- Playwright: `e2e/dashboard.spec.ts`, `e2e/agent-detail.spec.ts`.

## Required Env Vars
| Var                    | Default                       |
| ---------------------- | ----------------------------- |
| `NEXT_PUBLIC_BACKEND_URL` | `http://localhost:8080`    |
| `NEXT_PUBLIC_WS_URL`       | `ws://localhost:8080/v1/stream` |
