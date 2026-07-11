"""pytest config."""
import sys
from pathlib import Path

# Ensure ``ecomatrix`` package is importable when running tests from apps/agent/.
ROOT = Path(__file__).resolve().parent.parent
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))
