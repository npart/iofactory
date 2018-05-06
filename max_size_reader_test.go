package iofactory

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxSizeReader(t *testing.T) {
	bufSize := 1024
	readSize := 20
	buf := make([]byte, bufSize)
	reader := NewMaxSizeReader(bytes.NewReader(buf), readSize)

	maxN := 0
	totalRead := 0
	for {
		buf2 := make([]byte, bufSize)
		n, err := reader.Read(buf2)
		maxN = maxInt(maxN, n)
		totalRead += n
		if err != nil {
			break
		}
	}

	assert.Equal(t, readSize, maxN)
	assert.Equal(t, totalRead, bufSize)
}
