"use client";

import { ReactNode, useRef } from "react";
import { motion, useMotionTemplate, useMotionValue, useSpring } from "framer-motion";

interface Props {
  children: ReactNode;
  className?: string;
  tone?: "cyan" | "gold";
}

const RING: Record<NonNullable<Props["tone"]>, string> = {
  cyan: "ring-cyan-glow",
  gold: "ring-gold-glow",
};

// Aceternity 3D-card-lite: subtle perspective tilt + ring on hover, gold ring on focus.
export function ThreeDCard({ children, className, tone = "cyan" }: Props) {
  const ref = useRef<HTMLDivElement | null>(null);
  const rx = useSpring(0, { stiffness: 220, damping: 22 });
  const ry = useSpring(0, { stiffness: 220, damping: 22 });

  const onMove = (e: React.MouseEvent) => {
    const el = ref.current;
    if (!el) return;
    const r = el.getBoundingClientRect();
    const px = (e.clientX - r.left) / r.width - 0.5;   // -0.5..0.5
    const py = (e.clientY - r.top) / r.height - 0.5;
    ry.set(px * 6);                                     // up to 6° yaw
    rx.set(-py * 6);                                    // up to 6° pitch
  };
  const onLeave = () => {
    rx.set(0);
    ry.set(0);
  };

  const transform = useMotionTemplate`perspective(800px) rotateX(${rx}deg) rotateY(${ry}deg)`;

  return (
    <motion.div
      ref={ref}
      onMouseMove={onMove}
      onMouseLeave={onLeave}
      style={{ transform }}
      className={
        "rounded-md border border-hairline bg-panel p-4 transition-shadow duration-150 " +
        "hover:" + RING[tone] + " " + (className ?? "")
      }
    >
      {children}
    </motion.div>
  );
}
