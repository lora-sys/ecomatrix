"use client";

/**
 * ISS-FRONTEND: a single fixed ambient layer that anchors every page.
 * Three drifting radial blobs (cyan, violet, gold), a slow scanline
 * overlay, and a film-grain SVG rasterizer. All non-blocking
 * (pointer-events: none) so the rest of the UI stays clickable.
 */
import { memo } from "react";

export const AmbientBackground = memo(function AmbientBackground() {
  return (
    <div
      aria-hidden
      className="pointer-events-none fixed inset-0 -z-10 overflow-hidden"
    >
      {/* Three drifting blobs. */}
      <div
        className="ecomatrix-blob absolute -left-40 -top-40 h-[60vmin] w-[60vmin] rounded-full"
        style={{
          background:
            "radial-gradient(circle at 30% 30%, rgba(34,211,238,0.45) 0%, rgba(34,211,238,0) 65%)",
        }}
      />
      <div
        className="ecomatrix-blob absolute right-[-20%] top-[30%] h-[55vmin] w-[55vmin] rounded-full"
        style={{
          background:
            "radial-gradient(circle at 60% 40%, rgba(124,92,255,0.45) 0%, rgba(124,92,255,0) 65%)",
          animationDelay: "-6s",
        }}
      />
      <div
        className="ecomatrix-blob absolute bottom-[-20%] left-[20%] h-[55vmin] w-[55vmin] rounded-full"
        style={{
          background:
            "radial-gradient(circle at 50% 50%, rgba(245,192,68,0.32) 0%, rgba(245,192,68,0) 65%)",
          animationDelay: "-12s",
        }}
      />

      {/* Scanlines + grain — both at 1x CSS pixel size so it stays cheap. */}
      <div className="ecomatrix-scanlines absolute inset-0" />
      <div className="ecomatrix-grain absolute inset-0 mix-blend-screen" />

      {/* Subtle equator line for compositional depth. */}
      <div
        className="absolute left-0 right-0 top-1/2 h-px"
        style={{
          background:
            "linear-gradient(90deg, transparent 0%, rgba(125,211,252,0.18) 50%, transparent 100%)",
        }}
      />
    </div>
  );
});
