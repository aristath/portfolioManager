"""Router Bridge client for Arduino Uno Q.

Provides a Python interface to call functions exposed by the MCU sketch
via Router Bridge (msgpack RPC over Unix socket).
"""

import logging
import socket
import struct
from typing import Any, Optional

logger = logging.getLogger(__name__)

# Router Bridge socket path
ROUTER_BRIDGE_SOCKET = "/var/run/arduino-router.sock"

try:
    import msgpack
except ImportError:
    msgpack = None
    logger.warning("msgpack not available - Router Bridge client will not work")


class RouterBridgeClient:
    """Client for communicating with MCU via Router Bridge."""

    def __init__(self, socket_path: str = ROUTER_BRIDGE_SOCKET):
        """Initialize Router Bridge client.

        Args:
            socket_path: Path to Router Bridge Unix socket
        """
        self.socket_path = socket_path
        if msgpack is None:
            raise ImportError("msgpack package is required for Router Bridge client")

    def _connect(self) -> socket.socket:
        """Create and return a connected socket to Router Bridge."""
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        try:
            sock.connect(self.socket_path)
            return sock
        except Exception as e:
            raise ConnectionError(f"Failed to connect to Router Bridge at {self.socket_path}: {e}") from e

    def call(self, function_name: str, *args: Any, timeout: float = 5.0) -> Optional[Any]:
        """Call a function on the MCU via Router Bridge.

        Args:
            function_name: Name of the function to call (must be registered via Bridge.provide)
            *args: Arguments to pass to the function
            timeout: Timeout in seconds for the RPC call

        Returns:
            Function return value, or None if function returns void

        Raises:
            ConnectionError: If connection to Router Bridge fails
            TimeoutError: If call times out
            RuntimeError: If RPC call fails
        """
        if msgpack is None:
            raise ImportError("msgpack package is required")

        sock = self._connect()
        try:
            sock.settimeout(timeout)

            # Router Bridge RPC protocol: [message_id, function_name, args]
            # Message ID is a 32-bit integer (we use 1 for simplicity)
            message = [1, function_name, list(args)]

            # Pack message with msgpack
            packed = msgpack.packb(message, use_bin_type=True)

            # Send length prefix (4 bytes, big-endian)
            length = len(packed)
            sock.sendall(struct.pack(">I", length))

            # Send message
            sock.sendall(packed)

            # Receive response length (4 bytes, big-endian)
            length_bytes = sock.recv(4)
            if len(length_bytes) != 4:
                raise RuntimeError("Invalid response from Router Bridge (no length)")

            response_length = struct.unpack(">I", length_bytes)[0]

            # Receive response data
            response_data = b""
            while len(response_data) < response_length:
                chunk = sock.recv(response_length - len(response_data))
                if not chunk:
                    raise RuntimeError("Incomplete response from Router Bridge")
                response_data += chunk

            # Unpack response
            response = msgpack.unpackb(response_data, raw=False)

            # Router Bridge response format: [message_id, error, result]
            if len(response) != 3:
                raise RuntimeError(f"Invalid response format: {response}")

            msg_id, error, result = response

            if error is not None:
                raise RuntimeError(f"Router Bridge error: {error}")

            return result

        except socket.timeout as e:
            raise TimeoutError(f"Router Bridge call timed out after {timeout}s") from e
        except Exception as e:
            logger.error(f"Router Bridge call failed: {e}")
            raise
        finally:
            sock.close()


# Singleton instance
_client: Optional[RouterBridgeClient] = None


def get_client() -> RouterBridgeClient:
    """Get or create singleton Router Bridge client instance."""
    global _client
    if _client is None:
        _client = RouterBridgeClient()
    return _client


def call(function_name: str, *args: Any, timeout: float = 5.0) -> Optional[Any]:
    """Convenience function to call Router Bridge function.

    Args:
        function_name: Name of the function to call
        *args: Arguments to pass to the function
        timeout: Timeout in seconds

    Returns:
        Function return value, or None if function returns void
    """
    return get_client().call(function_name, *args, timeout=timeout)
