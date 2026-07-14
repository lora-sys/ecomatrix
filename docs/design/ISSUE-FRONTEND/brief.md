# ISS-FRONTEND — Creative Upgrade Brief

**Trigger**: User asked to upgrade the EcoMatrix dashboard to "Awwwards-level" and to place screenshots in the README.

**Theme**: Cyber command-deck (keeps the project's existing visual language but tightens it). Dark canvas (near-black with a faint cyan/violet gradient haze), monospaced telemetry font for numerals + labels, display font for hero phrases, glow borders on every primary card.

**Big swing**:
1. Add an animated ambient background (radial gradient blobs + scanline + grain) that frames every page.
2. Replace the small "上帝视角 / EcoMatrix 经济演化网络的实时观测" header with a layered masthead: a tiny kicker, a giant display headline (`ECOMATRIX` + 上帝视角), and a running ticker ribbon that scrolls live metrics.
3. KPI tiles rebuilt as full-bleed, large numeral "BigMetric" cards with hover/wave animations.
4. Body grid becomes intentionally asymmetric (a hero rail that bleeds out, then nested tiles).
5. Agent page becomes a dossier: huge portrait header + telemetry sidecar.
6. Supervisor detail page becomes a transmissal log header.

**Stack constraint**: keep the existing dep set (Next.js 15, framer-motion, tailwind). No GSAP, no Three.js. CSS gradients + framer-motion + small SVG layers.

**Acceptance**:
- Before/after screenshots captured at desktop + mobile, saved to `docs/design/ISSUE-FRONTEND/{before,after}/`.
- `npm run typecheck && npm run lint && npm run build` clean.
- The existing Playwright suite still passes (panels still visible, dashboard still renders).
- README gets a "Demo" section embedding the before/after pair.
