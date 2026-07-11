#!/usr/bin/env bash
# Regenerate docs/.index/{manifest,relations,freshness}.json
# Phase-1 stub: emits a flat manifest. Real indexer lives in a later Issue.
set -euo pipefail

cd "$(dirname "$0")/.."

mkdir -p docs/.index

python3 - <<'PY'
import hashlib, json, os, time
from pathlib import Path

root = Path(".")
docs = root / "docs"
out = docs / ".index"
out.mkdir(exist_ok=True)

manifest = []
for p in sorted(docs.rglob("*.md")):
    rel = p.relative_to(root).as_posix()
    h = hashlib.sha256(p.read_bytes()).hexdigest()[:12]
    manifest.append({"path": rel, "sha256_12": h, "size": p.stat().st_size})

(out / "manifest.json").write_text(json.dumps(manifest, indent=2))
(out / "relations.json").write_text(json.dumps({"edges": []}, indent=2))
(out / "freshness.json").write_text(json.dumps({"generated_at": int(time.time()), "count": len(manifest)}, indent=2))
print(f"Indexed {len(manifest)} markdown files.")
PY
