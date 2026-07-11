// Value damping for live KPI tiles.
// Returns a value that smoothly approaches the target over `duration` ms.

export function dampValue(prev: number, target: number, dtMs: number, tau = 300): number {
  if (prev === target) return target;
  const alpha = 1 - Math.exp(-dtMs / tau);
  return prev + (target - prev) * alpha;
}

export function clamp(n: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, n));
}
