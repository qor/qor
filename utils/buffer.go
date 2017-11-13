package utils

import (
	"bytes"
)

// ClosingBuffer implement Closer interface for Buffer
type ClosingBuffer struct {
	*bytes.Buffer
}

// Close implement Closer interface for Buffer
func (ClosingBuffer) Close() error {
	return nil
}
