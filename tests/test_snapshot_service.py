import math
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from sentinel.snapshot_service import _format_progress


def test_format_progress_includes_percent_and_eta():
    text = _format_progress(start_ts=0.0, current_idx=50, total=200, date_str="2026-02-05", now_ts=100.0)
    assert "50/200" in text
    assert "25.0%" in text
    assert "eta" in text
    assert "2026-02-05" in text
    assert "elapsed" in text
    assert "s" in text
    assert not math.isnan(float(text.split("eta=")[1].split("s")[0]))
