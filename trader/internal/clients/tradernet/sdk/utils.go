package sdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// sign generates a SHA256 HMAC signature of the message using the key.
// This matches the Python SDK's sign function:
// hmac.new(key.encode(), msg.encode(), digestmod='sha256').hexdigest()
func sign(key, message string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// stringify converts data to a compact JSON string (no spaces).
// This matches Python's json.dumps(data, separators=(',', ':'))
// CRITICAL: For structs, field order is preserved (matches Python dict insertion order)
// CRITICAL: For maps, key order may vary (use structs for deterministic order)
func stringify(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
