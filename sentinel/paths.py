"""Central data directory configuration."""

import os
from pathlib import Path

# Project root is the parent of the sentinel package directory
_PROJECT_ROOT = Path(__file__).parent.parent

DATA_DIR = Path(os.environ.get("SENTINEL_DATA_DIR", _PROJECT_ROOT / "data"))
