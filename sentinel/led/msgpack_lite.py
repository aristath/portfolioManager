"""Tiny MessagePack subset for UNO Q arduino-router RPC.

We avoid external dependencies because Sentinel often runs in constrained
environments. This implements only what we need for MsgPack-RPC:
  - nil, bool
  - int/uint (up to 64-bit)
  - float64
  - str (utf-8)
  - bin (bytes)
  - array (list/tuple)
  - map (dict)

It is not a full MessagePack implementation.
"""

from __future__ import annotations

import struct
from typing import Any, Iterator


def packb(obj: Any) -> bytes:
    out = bytearray()
    _pack_into(out, obj)
    return bytes(out)


def _pack_into(out: bytearray, obj: Any) -> None:
    if obj is None:
        out.append(0xC0)
        return
    if obj is False:
        out.append(0xC2)
        return
    if obj is True:
        out.append(0xC3)
        return

    if isinstance(obj, int) and not isinstance(obj, bool):
        _pack_int(out, obj)
        return

    if isinstance(obj, float):
        out.append(0xCB)  # float64
        out.extend(struct.pack(">d", obj))
        return

    if isinstance(obj, str):
        b = obj.encode("utf-8")
        n = len(b)
        if n < 32:
            out.append(0xA0 | n)
        elif n < 256:
            out.extend((0xD9, n))
        elif n < 65536:
            out.append(0xDA)
            out.extend(struct.pack(">H", n))
        else:
            out.append(0xDB)
            out.extend(struct.pack(">I", n))
        out.extend(b)
        return

    if isinstance(obj, (bytes, bytearray, memoryview)):
        b = bytes(obj)
        n = len(b)
        if n < 256:
            out.extend((0xC4, n))
        elif n < 65536:
            out.append(0xC5)
            out.extend(struct.pack(">H", n))
        else:
            out.append(0xC6)
            out.extend(struct.pack(">I", n))
        out.extend(b)
        return

    if isinstance(obj, (list, tuple)):
        n = len(obj)
        if n < 16:
            out.append(0x90 | n)
        elif n < 65536:
            out.append(0xDC)
            out.extend(struct.pack(">H", n))
        else:
            out.append(0xDD)
            out.extend(struct.pack(">I", n))
        for it in obj:
            _pack_into(out, it)
        return

    if isinstance(obj, dict):
        n = len(obj)
        if n < 16:
            out.append(0x80 | n)
        elif n < 65536:
            out.append(0xDE)
            out.extend(struct.pack(">H", n))
        else:
            out.append(0xDF)
            out.extend(struct.pack(">I", n))
        for k, v in obj.items():
            _pack_into(out, k)
            _pack_into(out, v)
        return

    raise TypeError(f"msgpack_lite: unsupported type: {type(obj)!r}")


def _pack_int(out: bytearray, n: int) -> None:
    if 0 <= n <= 0x7F:
        out.append(n)  # positive fixint
    elif -32 <= n < 0:
        out.append(0xE0 | (n + 32))  # negative fixint
    elif 0 <= n <= 0xFF:
        out.extend((0xCC, n))
    elif 0 <= n <= 0xFFFF:
        out.append(0xCD)
        out.extend(struct.pack(">H", n))
    elif 0 <= n <= 0xFFFFFFFF:
        out.append(0xCE)
        out.extend(struct.pack(">I", n))
    elif 0 <= n <= 0xFFFFFFFFFFFFFFFF:
        out.append(0xCF)
        out.extend(struct.pack(">Q", n))
    elif -0x80 <= n < 0:
        out.append(0xD0)
        out.extend(struct.pack(">b", n))
    elif -0x8000 <= n < 0:
        out.append(0xD1)
        out.extend(struct.pack(">h", n))
    elif -0x80000000 <= n < 0:
        out.append(0xD2)
        out.extend(struct.pack(">i", n))
    elif -0x8000000000000000 <= n < 0:
        out.append(0xD3)
        out.extend(struct.pack(">q", n))
    else:
        raise OverflowError("msgpack_lite: int out of range")


class Unpacker:
    def __init__(self):
        self._buf = bytearray()
        self._off = 0

    def feed(self, data: bytes) -> None:
        if data:
            self._buf.extend(data)

    def __iter__(self) -> Iterator[Any]:
        while True:
            res = _unpack_one(self._buf, self._off)
            if res is None:
                # Compact buffer if we've consumed some.
                if self._off > 0:
                    del self._buf[: self._off]
                    self._off = 0
                return
            obj, new_off = res
            self._off = new_off
            yield obj


def _need(buf: bytearray, off: int, n: int) -> bool:
    return len(buf) - off < n


def _unpack_one(buf: bytearray, off: int) -> tuple[Any, int] | None:
    if _need(buf, off, 1):
        return None
    b0 = buf[off]
    off += 1

    # Positive fixint
    if b0 <= 0x7F:
        return b0, off
    # Fixmap
    if 0x80 <= b0 <= 0x8F:
        n = b0 & 0x0F
        m: dict[Any, Any] = {}
        for _ in range(n):
            k_res = _unpack_one(buf, off)
            if k_res is None:
                return None
            k, off = k_res
            v_res = _unpack_one(buf, off)
            if v_res is None:
                return None
            v, off = v_res
            m[k] = v
        return m, off
    # Fixarray
    if 0x90 <= b0 <= 0x9F:
        n = b0 & 0x0F
        arr: list[Any] = []
        for _ in range(n):
            it_res = _unpack_one(buf, off)
            if it_res is None:
                return None
            it, off = it_res
            arr.append(it)
        return arr, off
    # Fixstr
    if 0xA0 <= b0 <= 0xBF:
        n = b0 & 0x1F
        if _need(buf, off, n):
            return None
        s = bytes(buf[off : off + n]).decode("utf-8", errors="replace")
        return s, off + n
    # Negative fixint
    if b0 >= 0xE0:
        return b0 - 256, off

    # Nil / bool
    if b0 == 0xC0:
        return None, off
    if b0 == 0xC2:
        return False, off
    if b0 == 0xC3:
        return True, off

    # Bin
    if b0 == 0xC4:
        if _need(buf, off, 1):
            return None
        n = buf[off]
        off += 1
        if _need(buf, off, n):
            return None
        return bytes(buf[off : off + n]), off + n
    if b0 == 0xC5:
        if _need(buf, off, 2):
            return None
        (n,) = struct.unpack(">H", bytes(buf[off : off + 2]))
        off += 2
        if _need(buf, off, n):
            return None
        return bytes(buf[off : off + n]), off + n
    if b0 == 0xC6:
        if _need(buf, off, 4):
            return None
        (n,) = struct.unpack(">I", bytes(buf[off : off + 4]))
        off += 4
        if _need(buf, off, n):
            return None
        return bytes(buf[off : off + n]), off + n

    # Float64
    if b0 == 0xCA:
        if _need(buf, off, 4):
            return None
        (v,) = struct.unpack(">f", bytes(buf[off : off + 4]))
        return float(v), off + 4
    if b0 == 0xCB:
        if _need(buf, off, 8):
            return None
        (v,) = struct.unpack(">d", bytes(buf[off : off + 8]))
        return float(v), off + 8

    # Uint
    if b0 == 0xCC:
        if _need(buf, off, 1):
            return None
        return buf[off], off + 1
    if b0 == 0xCD:
        if _need(buf, off, 2):
            return None
        (v,) = struct.unpack(">H", bytes(buf[off : off + 2]))
        return v, off + 2
    if b0 == 0xCE:
        if _need(buf, off, 4):
            return None
        (v,) = struct.unpack(">I", bytes(buf[off : off + 4]))
        return v, off + 4
    if b0 == 0xCF:
        if _need(buf, off, 8):
            return None
        (v,) = struct.unpack(">Q", bytes(buf[off : off + 8]))
        return v, off + 8

    # Int
    if b0 == 0xD0:
        if _need(buf, off, 1):
            return None
        (v,) = struct.unpack(">b", bytes(buf[off : off + 1]))
        return v, off + 1
    if b0 == 0xD1:
        if _need(buf, off, 2):
            return None
        (v,) = struct.unpack(">h", bytes(buf[off : off + 2]))
        return v, off + 2
    if b0 == 0xD2:
        if _need(buf, off, 4):
            return None
        (v,) = struct.unpack(">i", bytes(buf[off : off + 4]))
        return v, off + 4
    if b0 == 0xD3:
        if _need(buf, off, 8):
            return None
        (v,) = struct.unpack(">q", bytes(buf[off : off + 8]))
        return v, off + 8

    # Str
    if b0 == 0xD9:
        if _need(buf, off, 1):
            return None
        n = buf[off]
        off += 1
        if _need(buf, off, n):
            return None
        s = bytes(buf[off : off + n]).decode("utf-8", errors="replace")
        return s, off + n
    if b0 == 0xDA:
        if _need(buf, off, 2):
            return None
        (n,) = struct.unpack(">H", bytes(buf[off : off + 2]))
        off += 2
        if _need(buf, off, n):
            return None
        s = bytes(buf[off : off + n]).decode("utf-8", errors="replace")
        return s, off + n
    if b0 == 0xDB:
        if _need(buf, off, 4):
            return None
        (n,) = struct.unpack(">I", bytes(buf[off : off + 4]))
        off += 4
        if _need(buf, off, n):
            return None
        s = bytes(buf[off : off + n]).decode("utf-8", errors="replace")
        return s, off + n

    # Array / map
    if b0 == 0xDC:
        if _need(buf, off, 2):
            return None
        (n,) = struct.unpack(">H", bytes(buf[off : off + 2]))
        off += 2
        arr: list[Any] = []
        for _ in range(n):
            it_res = _unpack_one(buf, off)
            if it_res is None:
                return None
            it, off = it_res
            arr.append(it)
        return arr, off
    if b0 == 0xDD:
        if _need(buf, off, 4):
            return None
        (n,) = struct.unpack(">I", bytes(buf[off : off + 4]))
        off += 4
        arr: list[Any] = []
        for _ in range(n):
            it_res = _unpack_one(buf, off)
            if it_res is None:
                return None
            it, off = it_res
            arr.append(it)
        return arr, off

    if b0 == 0xDE:
        if _need(buf, off, 2):
            return None
        (n,) = struct.unpack(">H", bytes(buf[off : off + 2]))
        off += 2
        m: dict[Any, Any] = {}
        for _ in range(n):
            k_res = _unpack_one(buf, off)
            if k_res is None:
                return None
            k, off = k_res
            v_res = _unpack_one(buf, off)
            if v_res is None:
                return None
            v, off = v_res
            m[k] = v
        return m, off
    if b0 == 0xDF:
        if _need(buf, off, 4):
            return None
        (n,) = struct.unpack(">I", bytes(buf[off : off + 4]))
        off += 4
        m: dict[Any, Any] = {}
        for _ in range(n):
            k_res = _unpack_one(buf, off)
            if k_res is None:
                return None
            k, off = k_res
            v_res = _unpack_one(buf, off)
            if v_res is None:
                return None
            v, off = v_res
            m[k] = v
        return m, off

    raise ValueError(f"msgpack_lite: unsupported type byte 0x{b0:02x}")
