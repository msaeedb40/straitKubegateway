// Package tracing provides distributed trace context generation and propagation.
package tracing

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateTraceID creates a new 128-bit trace ID in hex format.
func GenerateTraceID() string {
	var bytes [16]byte
	_, _ = rand.Read(bytes[:])
	return hex.EncodeToString(bytes[:])
}

// GenerateSpanID creates a new 64-bit span ID in hex format.
func GenerateSpanID() string {
	var bytes [8]byte
	_, _ = rand.Read(bytes[:])
	return hex.EncodeToString(bytes[:])
}

// GenerateRequestID creates a formatted request ID (e.g. "req-abc12345").
func GenerateRequestID() string {
	var bytes [6]byte
	_, _ = rand.Read(bytes[:])
	return fmt.Sprintf("req-%s", hex.EncodeToString(bytes[:]))
}

// GenerateCommandID creates a formatted command ID (e.g. "cmd-abc12345").
func GenerateCommandID() string {
	var bytes [6]byte
	_, _ = rand.Read(bytes[:])
	return fmt.Sprintf("cmd-%s", hex.EncodeToString(bytes[:]))
}
