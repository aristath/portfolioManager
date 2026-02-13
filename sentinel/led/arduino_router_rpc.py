"""Minimal MsgPack-RPC client/server helpers for Arduino UNO Q arduino-router.

The UNO Q runs arduino-router as a MsgPack-RPC router over a UNIX socket
(`/var/run/arduino-router.sock` by default). MCU sketches use Arduino_RouterBridge
to call methods through the router; Linux-side processes can register and serve
methods by calling `$/register`.
"""

from __future__ import annotations

import socket
from dataclasses import dataclass
from typing import Any, Callable, Final

from sentinel.led import msgpack_lite as msgpack

REQUEST: Final[int] = 0
RESPONSE: Final[int] = 1
NOTIFY: Final[int] = 2


@dataclass(frozen=True)
class RpcError(Exception):
    code: int
    message: str

    def __str__(self) -> str:  # pragma: no cover
        return f"RpcError(code={self.code}, message={self.message})"


class UnixMsgpackRpc:
    """A small, blocking MsgPack-RPC connection over a UNIX stream socket."""

    def __init__(self, sock_path: str):
        self._sock_path = sock_path
        self._sock: socket.socket | None = None
        self._unpacker: msgpack.Unpacker | None = None
        self._next_id = 1

    def connect(self) -> None:
        s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        s.connect(self._sock_path)
        s.settimeout(1.0)
        self._sock = s
        self._unpacker = msgpack.Unpacker()

    def close(self) -> None:
        if self._sock is not None:
            try:
                self._sock.close()
            finally:
                self._sock = None
                self._unpacker = None

    def _recv_one(self) -> Any:
        if self._sock is None or self._unpacker is None:
            raise RuntimeError("RPC socket not connected")

        while True:
            for obj in self._unpacker:
                return obj
            data = self._sock.recv(4096)
            if not data:
                raise ConnectionError("RPC socket closed")
            self._unpacker.feed(data)

    def call(self, method: str, *params: Any) -> Any:
        if self._sock is None:
            raise RuntimeError("RPC socket not connected")

        msgid = self._next_id
        self._next_id += 1

        req = [REQUEST, msgid, method, list(params)]
        self._sock.sendall(msgpack.packb(req))

        # Wait for matching response id.
        while True:
            msg = self._recv_one()
            if not isinstance(msg, list) or len(msg) != 4:
                continue
            if msg[0] != RESPONSE or msg[1] != msgid:
                continue
            err = msg[2]
            if err is not None:
                # Convention: [code, message]
                if isinstance(err, list) and len(err) >= 2:
                    raise RpcError(int(err[0]), str(err[1]))
                raise RpcError(0xFF, str(err))
            return msg[3]


def serve_forever(
    rpc: UnixMsgpackRpc,
    methods: dict[str, Callable[[list[Any]], Any]],
) -> None:
    """Serve incoming requests forever on a connected rpc socket.

    `methods` values accept the params list and return a result.
    """
    if rpc._sock is None or rpc._unpacker is None:  # noqa: SLF001
        raise RuntimeError("RPC socket not connected")

    while True:
        msg = rpc._recv_one()
        if not isinstance(msg, list) or len(msg) not in (3, 4):
            continue

        msgtype = msg[0]
        if msgtype == REQUEST:
            msgid = msg[1]
            method = msg[2]
            params = msg[3] if len(msg) == 4 else []

            if not isinstance(method, str) or not isinstance(params, list):
                continue

            try:
                handler = methods[method]
                result = handler(params)
                resp = [RESPONSE, msgid, None, result]
            except KeyError:
                resp = [RESPONSE, msgid, [0xFF, f"unknown method: {method}"], None]
            except Exception as e:  # pragma: no cover
                resp = [RESPONSE, msgid, [0xFF, str(e)], None]

            rpc._sock.sendall(msgpack.packb(resp))  # noqa: SLF001

        elif msgtype == NOTIFY:
            # Ignore notifications for now.
            continue
        else:
            # Ignore responses (we're acting as a server).
            continue
