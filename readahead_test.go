package iofactory

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadaheadTypical(t *testing.T) {
	testReadahead(t, 1<<16, 4, 60000, 10)
}

func TestReadaheadLargeBufferSmallReadSize(t *testing.T) {
	testReadahead(t, 1<<20, 4, 8000, 5)
}

func TestReadaheadSingleByte(t *testing.T) {
	testReadahead(t, 1, 1, 1, 1024)
}

func TestReadaheadLarge(t *testing.T) {
	testReadahead(t, 1<<24, 4, 1000000, 8)
}

func testReadahead(t *testing.T, length, buffers, readSize, iterations int) {

	buf := RandomByteSlice(length * buffers * iterations) // Require recycling buffers a few times
	assert.NotNil(t, buf)

	randReader := NewMaxSizeReader(NewRandomSizeReader(bytes.NewReader(buf)), readSize)
	readahead, err := NewReadahead(randReader, buffers, length, readSize)
	assert.Nil(t, err)

	bufCopied := make([]byte, len(buf)+100)
	n, err := io.ReadFull(readahead, bufCopied)
	assert.True(t, strings.Contains(err.Error(), "EOF"))

	// Compare the buffers
	assert.True(t, len(buf) < len(bufCopied))
	assert.Equal(t, len(buf), length*buffers*iterations)
	assert.Equal(t, n, len(buf))
	assert.True(t, bytes.Equal(buf, bufCopied[:len(buf)]))
}

func BenchmarkReadaheadTypical(b *testing.B) {
	benchmarkReadahead(b, 1<<16, 4, 60000, 10)
}

func BenchmarkReadaheadLargeBufferSmallReadSize(b *testing.B) {
	benchmarkReadahead(b, 1<<20, 4, 8000, 5)
}

func BenchmarkReadaheadSingleByte(b *testing.B) {
	benchmarkReadahead(b, 1, 1, 1, 1024)
}

func BenchmarkReadaheadLarge(b *testing.B) {
	benchmarkReadahead(b, 1<<24, 4, 1000000, 8)
}

func benchmarkReadahead(b *testing.B, length, buffers, readSize, iterations int) {

	buf := RandomByteSlice(length * buffers * iterations) // Require recycling buffers a few times
	bufCopied := make([]byte, len(buf))

	b.Logf("Length %v, Buffers %v, Iterations %v -> %v MB", length, buffers, iterations, float64(len(buf))/(1024*1024))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reader := NewMaxSizeReader(bytes.NewReader(buf), readSize)
		readahead, _ := NewReadahead(reader, buffers, length, readSize)
		n, err := io.ReadFull(readahead, bufCopied)
		panicIf(n < len(buf) || err != nil)
	}
}
