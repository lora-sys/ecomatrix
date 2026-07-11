"use client";

import { useStore } from "../hooks/store";

export function ErrorBanner() {
  const fetchError = useStore((s) => s.fetchError);
  if (!fetchError) return null;

  return (
    <div
      role="alert"
      aria-live="assertive"
      className="border-b border-rose-900/60 bg-rose-950/30 px-6 py-2 font-mono text-xs text-accent-rose"
    >
      <div className="mx-auto flex max-w-[1440px] items-center justify-between gap-4">
        <span>
          <strong>后端连接异常：</strong>
          {fetchError}
        </span>
        <button
          type="button"
          onClick={() => useStore.getState().setFetchError(null)}
          className="rounded border border-rose-800/60 px-2 py-0.5 hover:bg-rose-900/30"
        >
          关闭
        </button>
      </div>
    </div>
  );
}
