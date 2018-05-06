package iofactory

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomSizeReader(t *testing.T) {
	bufSize := 1 << 20
	maxReadSize := 32
	totalRead := 0

	buf := make([]byte, bufSize)
	reader := NewRandomSizeReader(bytes.NewReader(buf))
	hitCounts := make([]int, maxReadSize+1)

	for {
		buf2 := make([]byte, maxReadSize)
		n, err := reader.Read(buf2)

		assert.True(t, n >= 0 && n <= maxReadSize)
		totalRead += n

		hitCounts[n]++

		if err != nil {
			break
		}
	}

	// This test should return 0 exactly once when the stream is completed.
	assert.Equal(t, hitCounts[0], 1)

	for i := 1; i <= maxReadSize; i++ {
		// Expecting read sizes for all values between 1 and maxReadSize
		assert.True(t, hitCounts[i] > 0)
	}
	assert.Equal(t, totalRead, bufSize)
}
