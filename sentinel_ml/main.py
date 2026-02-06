#!/usr/bin/env python3
"""Entry point for Sentinel ML service."""

import argparse

import uvicorn


def main():
    parser = argparse.ArgumentParser(description="Sentinel ML Service")
    parser.add_argument("--host", default="::", help="Host")
    parser.add_argument("--port", type=int, default=8001, help="Port")
    args = parser.parse_args()

    uvicorn.run("sentinel_ml.app:app", host=args.host, port=args.port, log_level="info")


if __name__ == "__main__":
    main()
