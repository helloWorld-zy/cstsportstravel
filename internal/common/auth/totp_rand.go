package auth

import (
	"crypto/rand"
	"io"
	"os"
)

// openURandom opens /dev/urandom or falls back to crypto/rand.Reader.
func openURandom() (io.ReadCloser, error) {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		// On Windows or when /dev/urandom is not available, use crypto/rand
		return io.NopCloser(rand.Reader), nil
	}
	return f, nil
}
