package utils

import (
	"bytes"
	"os"
)

// ClosingBuffer implement Closer interface for Buffer
type ClosingBuffer struct {
	*bytes.Buffer
}

// Close implement Closer interface for Buffer
func (cb *ClosingBuffer) Close() (err os.Error) {
	return
}
