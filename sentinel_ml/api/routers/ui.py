"""Minimal web UI for sentinel-ml operational visibility and settings."""
# ruff: noqa: E501

from __future__ import annotations

from urllib.parse import urlparse

from fastapi import APIRouter, Depends, HTTPException
from fastapi.responses import HTMLResponse
from pydantic import BaseModel
from typing_extensions import Annotated

from sentinel_ml.api.dependencies import CommonDependencies, get_common_deps
from sentinel_ml.clients.monolith_client import get_monolith_base_url, set_monolith_base_url

router = APIRouter(tags=["ui"])


class MonolithUrlUpdate(BaseModel):
    monolith_base_url: str


def _validate_base_url(url: str) -> str:
    value = url.strip().rstrip("/")
    parsed = urlparse(value)
    if parsed.scheme not in {"http", "https"} or not parsed.netloc:
        raise HTTPException(status_code=400, detail="monolith_base_url must be a valid http(s) URL")
    return value


@router.get("/", include_in_schema=False)
async def index() -> HTMLResponse:
    html = """
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Sentinel ML Console</title>
  <style>
    :root {
      --bg: #0f172a;
      --panel: #111827;
      --panel-2: #1f2937;
      --text: #e5e7eb;
      --muted: #9ca3af;
      --accent: #22c55e;
      --danger: #f97316;
      --border: #374151;
    }
    body { margin: 0; font-family: ui-sans-serif, system-ui, -apple-system, sans-serif; background: radial-gradient(circle at top, #0b1225, var(--bg)); color: var(--text); }
    .wrap { max-width: 960px; margin: 0 auto; padding: 24px; }
    h1 { margin: 0 0 18px; font-size: 28px; }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 12px; }
    .card { background: linear-gradient(180deg, var(--panel), var(--panel-2)); border: 1px solid var(--border); border-radius: 12px; padding: 14px; }
    .k { font-size: 12px; color: var(--muted); text-transform: uppercase; letter-spacing: 0.08em; }
    .v { margin-top: 6px; font-size: 26px; font-weight: 700; }
    .v.small { font-size: 16px; font-weight: 500; }
    .row { display: flex; gap: 10px; align-items: center; margin-top: 16px; }
    input { flex: 1; background: #0b1020; border: 1px solid var(--border); border-radius: 8px; color: var(--text); padding: 10px; }
    button { background: var(--accent); color: #052e16; border: 0; border-radius: 8px; padding: 10px 14px; font-weight: 700; cursor: pointer; }
    #status { margin-top: 10px; font-size: 13px; color: var(--muted); }
    #status.error { color: var(--danger); }
  </style>
</head>
<body>
  <div class="wrap">
    <h1>Sentinel ML Console</h1>
    <div class="grid" id="stats"></div>
    <div class="card" style="margin-top:12px;">
      <div class="k">Monolith Base URL</div>
      <div class="row">
        <input id="mono" type="text" placeholder="http://localhost:8000" />
        <button id="save">Save</button>
      </div>
      <div id="status"></div>
    </div>
  </div>
  <script>
    async function load() {
      const [cfg, status] = await Promise.all([
        fetch('/ui/config').then(r => r.json()),
        fetch('/ml/status').then(r => r.json())
      ]);
      document.getElementById('mono').value = cfg.monolith_base_url;

      const cards = [
        ['ML-enabled Securities', status.securities_ml_enabled ?? 0],
        ['Symbols With Models', status.symbols_with_models ?? 0],
        ['Training Samples', status.total_training_samples ?? 0],
        ['Monolith URL', cfg.monolith_base_url]
      ];
      document.getElementById('stats').innerHTML = cards.map(([k,v]) =>
        `<div class="card"><div class="k">${k}</div><div class="v ${typeof v === 'string' ? 'small' : ''}">${v}</div></div>`
      ).join('');
    }

    document.getElementById('save').addEventListener('click', async () => {
      const statusEl = document.getElementById('status');
      statusEl.className = '';
      statusEl.textContent = 'Saving...';
      try {
        const monolith_base_url = document.getElementById('mono').value;
        const res = await fetch('/ui/config', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ monolith_base_url })
        });
        if (!res.ok) {
          const body = await res.json();
          throw new Error(body.detail || 'Failed to save');
        }
        statusEl.textContent = 'Saved';
        await load();
      } catch (e) {
        statusEl.className = 'error';
        statusEl.textContent = e.message;
      }
    });

    load();
  </script>
</body>
</html>
"""
    return HTMLResponse(html)


@router.get("/ui/config")
async def get_ui_config(deps: Annotated[CommonDependencies, Depends(get_common_deps)]) -> dict[str, str]:
    saved = await deps.ml_db.get_service_setting("monolith_base_url")
    return {"monolith_base_url": saved or get_monolith_base_url()}


@router.put("/ui/config")
async def set_ui_config(
    payload: MonolithUrlUpdate,
    deps: Annotated[CommonDependencies, Depends(get_common_deps)],
) -> dict[str, str]:
    url = _validate_base_url(payload.monolith_base_url)
    await deps.ml_db.set_service_setting("monolith_base_url", url)
    set_monolith_base_url(url)
    return {"status": "ok", "monolith_base_url": url}
