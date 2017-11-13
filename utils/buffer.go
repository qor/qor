package utils

import "io"

// ClosingReadSeeker implement Closer interface for ReadSeeker
type ClosingReadSeeker struct {
	io.ReadSeeker
}

// Close implement Closer interface for Buffer
func (ClosingReadSeeker) Close() error {
	return nil
}
