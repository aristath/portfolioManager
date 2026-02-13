#!/usr/bin/env python3
"""UNO Q MPU-side server for the NeoPixel heatmap (official arduino-router bridge).

Sketch side (MCU):
  - calls Bridge.call("heatmap/get").result(out)

Linux side (MPU, this script):
  - connects to /var/run/arduino-router.sock
  - registers method name "heatmap/get" via $/register
  - serves MsgPack-RPC requests and responds with:
      [[before40...],[after40...]]
"""

from __future__ import annotations

import asyncio
import logging
import os
from typing import Any

from sentinel.database import Database
from sentinel.led.arduino_router_rpc import UnixMsgpackRpc, serve_forever
from sentinel.led.heatmap_parts import SecurityScore, build_sorted_parts, clamp_score
from sentinel.planner import Planner

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger("uno_q_heatmap_router_server")


async def compute_before_after() -> list[list[float]]:
    db = Database()
    await db.connect()
    planner = Planner()
    try:
        positions = await db.get_all_positions()
        if not positions:
            return [[0.0] * 40, [0.0] * 40]

        before_values: dict[str, float] = {}
        scores: dict[str, float] = {}
        for p in positions:
            sym = str(p["symbol"])
            qty = float(p.get("quantity") or 0.0)
            current_price = float(p.get("current_price") or 0.0)
            avg_cost = float(p.get("avg_cost") or 0.0)

            before_values[sym] = max(0.0, qty * current_price)
            if avg_cost > 0 and current_price > 0:
                pl = (current_price - avg_cost) / avg_cost
            else:
                pl = 0.0
            scores[sym] = clamp_score(pl, clamp_abs=0.5)

        total_before = sum(before_values.values())
        if total_before <= 0:
            return [[0.0] * 40, [0.0] * 40]

        after_values = dict(before_values)
        try:
            recs = await planner.get_recommendations()
        except Exception as e:
            logger.warning(f"Failed to get recommendations; after==before: {e}")
            recs = []

        for r in recs or []:
            sym = getattr(r, "symbol", None)
            if not sym:
                continue
            delta = float(getattr(r, "value_delta_eur", 0.0) or 0.0)
            after_values[sym] = max(0.0, after_values.get(sym, 0.0) + delta)

        total_after = sum(after_values.values()) or total_before

        before_scores: list[SecurityScore] = []
        after_scores: list[SecurityScore] = []
        for sym, score in scores.items():
            w_before = before_values.get(sym, 0.0) / total_before
            w_after = after_values.get(sym, 0.0) / total_after
            before_scores.append(SecurityScore(symbol=sym, weight=w_before, score=score))
            after_scores.append(SecurityScore(symbol=sym, weight=w_after, score=score))

        before40 = build_sorted_parts(before_scores, total_parts=40)
        after40 = build_sorted_parts(after_scores, total_parts=40)
        return [before40, after40]
    finally:
        await db.close()


def main() -> int:
    sock_path = os.environ.get("ARDUINO_ROUTER_SOCK", "/var/run/arduino-router.sock")
    method = os.environ.get("HEATMAP_METHOD", "heatmap/get")

    rpc = UnixMsgpackRpc(sock_path)
    rpc.connect()
    logger.info(f"Connected to arduino-router socket {sock_path}")

    # Register the method name so the router can route calls to this connection.
    rpc.call("$/register", method)
    logger.info(f"Registered method {method!r}")

    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)

    def handle_get(_: list[Any]) -> Any:
        # Compute async data in the background event loop but block this handler
        # (router calls are synchronous per request).
        return loop.run_until_complete(compute_before_after())

    try:
        serve_forever(rpc, {method: handle_get})
    finally:
        rpc.close()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
