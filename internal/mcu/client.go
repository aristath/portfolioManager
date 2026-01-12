// Package mcu provides a client for communicating with the Arduino MCU
// via the arduino-router Unix socket using msgpack-rpc protocol.
package mcu

import (
	"errors"
	"net"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	// DefaultSocketPath is the default path to the arduino-router Unix socket.
	DefaultSocketPath = "/var/run/arduino-router.sock"

	// RetryInterval is the time to wait between reconnection attempts.
	RetryInterval = 120 * time.Second

	// MaxRetries is the maximum number of reconnection attempts.
	MaxRetries = 120

	// ReadTimeout is the timeout for reading responses from the socket.
	ReadTimeout = 5 * time.Second

	// WriteTimeout is the timeout for writing to the socket.
	WriteTimeout = 5 * time.Second
)

var (
	// ErrNotConnected is returned when attempting to send without a connection.
	ErrNotConnected = errors.New("not connected to MCU")

	// ErrSocketNotFound is returned when the socket file doesn't exist.
	ErrSocketNotFound = errors.New("MCU socket not found")

	// ErrMaxRetriesExceeded is returned when max reconnection attempts are exceeded.
	ErrMaxRetriesExceeded = errors.New("max reconnection retries exceeded")
)

// Client manages the connection to the arduino-router Unix socket.
type Client struct {
	socketPath  string
	conn        net.Conn
	mu          sync.Mutex
	log         zerolog.Logger
	msgID       uint32
	retryCount  int
	isConnected bool
}

// NewClient creates a new MCU client.
// Returns nil if the socket doesn't exist (not running on Arduino hardware).
func NewClient(socketPath string, log zerolog.Logger) (*Client, error) {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}

	// Check if socket exists - graceful skip if not on Arduino hardware
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		log.Info().Str("socket_path", socketPath).Msg("MCU socket not found - display features disabled")
		return nil, ErrSocketNotFound
	}

	client := &Client{
		socketPath: socketPath,
		log:        log.With().Str("component", "mcu_client").Logger(),
	}

	// Attempt initial connection
	if err := client.connect(); err != nil {
		client.log.Warn().Err(err).Msg("Initial MCU connection failed, will retry on first call")
	}

	return client, nil
}

// connect establishes a connection to the Unix socket.
func (c *Client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.connectLocked()
}

// connectLocked establishes a connection (caller must hold mutex).
func (c *Client) connectLocked() error {
	// Close existing connection if any
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		c.isConnected = false
	}

	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		c.log.Debug().Err(err).Str("socket_path", c.socketPath).Msg("Failed to connect to MCU socket")
		return err
	}

	c.conn = conn
	c.isConnected = true
	c.retryCount = 0
	c.log.Info().Str("socket_path", c.socketPath).Msg("Connected to MCU socket")

	return nil
}

// reconnect attempts to reconnect with retry logic.
func (c *Client) reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.reconnectLocked()
}

// reconnectLocked attempts to reconnect (caller must hold mutex).
func (c *Client) reconnectLocked() error {
	for c.retryCount < MaxRetries {
		c.retryCount++
		c.log.Info().
			Int("attempt", c.retryCount).
			Int("max_retries", MaxRetries).
			Dur("retry_interval", RetryInterval).
			Msg("Attempting MCU reconnection")

		// Wait before retry (except on first attempt after disconnect)
		if c.retryCount > 1 {
			c.mu.Unlock()
			time.Sleep(RetryInterval)
			c.mu.Lock()
		}

		err := c.connectLocked()
		if err == nil {
			return nil
		}

		c.log.Warn().Err(err).Int("attempt", c.retryCount).Msg("MCU reconnection attempt failed")
	}

	c.log.Error().Int("max_retries", MaxRetries).Msg("Max MCU reconnection retries exceeded")
	return ErrMaxRetriesExceeded
}

// Close closes the connection to the MCU.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.isConnected = false
		c.log.Info().Msg("MCU connection closed")
		return err
	}
	return nil
}

// IsConnected returns whether the client is currently connected.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isConnected
}

// getConn returns the current connection, attempting to reconnect if needed.
// Caller must NOT hold the mutex.
func (c *Client) getConn() (net.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && c.isConnected {
		return c.conn, nil
	}

	// Try to reconnect
	if err := c.connectLocked(); err != nil {
		return nil, err
	}

	return c.conn, nil
}

// nextMsgID returns the next message ID for RPC calls.
func (c *Client) nextMsgID() uint32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.msgID++
	return c.msgID
}

// markDisconnected marks the connection as disconnected.
func (c *Client) markDisconnected() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isConnected = false
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
