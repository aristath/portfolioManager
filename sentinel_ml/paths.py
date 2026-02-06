"""Central data directory configuration for sentinel-ml."""

import os
from pathlib import Path

_PROJECT_ROOT = Path(__file__).parent.parent
DATA_DIR = Path(os.environ.get("SENTINEL_DATA_DIR", _PROJECT_ROOT / "data"))
