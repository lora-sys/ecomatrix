"use client";

import { useEffect, useRef } from "react";
import { useStore } from "./store";

const WS = process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8080/v1/stream";

export function useEconomatrixStream(): { connected: boolean } {
  const applyEvent = useStore((s) => s.applyEvent);
  const setConnected = useStore((s) => s.setConnected);
  const wsRef = useRef<WebSocket | null>(null);
  const retriesRef = useRef(0);

  useEffect(() => {
    let stopped = false;

    const connect = () => {
      if (stopped) return;
      try {
        const ws = new WebSocket(WS);
        wsRef.current = ws;
        ws.onopen = () => {
          retriesRef.current = 0;
          setConnected(true);
        };
        ws.onmessage = (ev) => {
          try {
            const data = JSON.parse(String(ev.data));
            applyEvent(data);
          } catch {
            /* ignore */
          }
        };
        ws.onerror = () => setConnected(false);
        ws.onclose = () => {
          setConnected(false);
          if (stopped) return;
          // Exponential backoff with jitter: 1s → 30s, ±20%.
          const base = Math.min(30000, 1000 * 2 ** retriesRef.current);
          const jitter = base * 0.2 * (Math.random() - 0.5) * 2;
          setTimeout(connect, base + jitter);
          retriesRef.current++;
        };
      } catch {
        setConnected(false);
      }
    };

    connect();
    return () => {
      stopped = true;
      wsRef.current?.close();
    };
  }, [applyEvent, setConnected]);

  return { connected: useStore((s) => s.connected) };
}
