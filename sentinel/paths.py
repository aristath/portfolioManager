"""Central data directory configuration."""

import os
from pathlib import Path

DATA_DIR = Path(os.environ.get("SENTINEL_DATA_DIR", Path.home() / "data"))
