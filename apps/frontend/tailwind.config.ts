import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        canvas: "#070A12",
        panel: "#0E1422",
        "panel-2": "#141B2D",
        hairline: "#1F2A44",
        ink: {
          DEFAULT: "#E6EDF7",
          muted: "#8A95B2",
          dim: "#5A6685",
        },
        accent: {
          cyan: "#22D3EE",
          violet: "#7C5CFF",
          gold: "#F5C044",
          rose: "#F43F5E",
          emerald: "#34D399",
        },
      },
      fontFamily: {
        display: ['var(--font-display)', "system-ui", "sans-serif"],
        body: ['var(--font-body)', "system-ui", "sans-serif"],
        mono: ['var(--font-mono)', "ui-monospace", "monospace"],
      },
      fontSize: {
        'display-xl': ['2.25rem', { lineHeight: '1.2', letterSpacing: '-0.02em' }],
      },
      boxShadow: {
        neon: "0 0 0 1px rgba(34,211,238,0.18), 0 0 32px rgba(34,211,238,0.10)",
        neonGold: "0 0 0 1px rgba(245,192,68,0.20), 0 0 32px rgba(245,192,68,0.10)",
      },
    },
  },
  plugins: [],
};
export default config;
