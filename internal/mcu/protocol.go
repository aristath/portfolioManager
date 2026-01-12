package mcu

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

// msgpack-rpc message types
const (
	msgTypeRequest      = 0
	msgTypeResponse     = 1
	msgTypeNotification = 2
)

// RPCError represents an error returned by the MCU.
type RPCError struct {
	Code    int
	Message string
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

// Call sends an RPC request and waits for a response.
func (c *Client) Call(method string, params ...interface{}) (interface{}, error) {
	conn, err := c.getConn()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	msgID := c.nextMsgID()

	// Build request: [type=0, msgid, method, params]
	request := []interface{}{msgTypeRequest, msgID, method, params}

	// Encode and send
	if err := c.sendMessage(conn, request); err != nil {
		c.markDisconnected()
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	response, err := c.readResponse(conn)
	if err != nil {
		c.markDisconnected()
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response: [type=1, msgid, error, result]
	if len(response) < 4 {
		return nil, fmt.Errorf("invalid response format: expected 4 elements, got %d", len(response))
	}

	respType, ok := toInt(response[0])
	if !ok || respType != msgTypeResponse {
		return nil, fmt.Errorf("unexpected response type: %v", response[0])
	}

	// Check error field
	if response[2] != nil {
		errData, ok := response[2].([]interface{})
		if ok && len(errData) >= 2 {
			code, _ := toInt(errData[0])
			msg, _ := errData[1].(string)
			return nil, &RPCError{Code: code, Message: msg}
		}
		return nil, fmt.Errorf("RPC error: %v", response[2])
	}

	return response[3], nil
}

// Notify sends an RPC notification (no response expected).
func (c *Client) Notify(method string, params ...interface{}) error {
	conn, err := c.getConn()
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	// Build notification: [type=2, method, params]
	notification := []interface{}{msgTypeNotification, method, params}

	// Encode and send
	if err := c.sendMessage(conn, notification); err != nil {
		c.markDisconnected()
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

// sendMessage encodes and sends a msgpack message.
func (c *Client) sendMessage(conn io.Writer, msg interface{}) error {
	// Set write deadline if supported
	if nc, ok := conn.(net.Conn); ok {
		nc.SetWriteDeadline(time.Now().Add(WriteTimeout))
	}

	encoder := msgpack.NewEncoder(conn)
	return encoder.Encode(msg)
}

// readResponse reads and decodes a msgpack response.
func (c *Client) readResponse(conn io.Reader) ([]interface{}, error) {
	// Set read deadline if supported
	if nc, ok := conn.(net.Conn); ok {
		nc.SetReadDeadline(time.Now().Add(ReadTimeout))
	}

	decoder := msgpack.NewDecoder(conn)
	var response []interface{}
	if err := decoder.Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}

// toInt converts various numeric types to int.
func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int8:
		return int(n), true
	case int16:
		return int(n), true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case uint:
		return int(n), true
	case uint8:
		return int(n), true
	case uint16:
		return int(n), true
	case uint32:
		return int(n), true
	case uint64:
		return int(n), true
	case float32:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}
