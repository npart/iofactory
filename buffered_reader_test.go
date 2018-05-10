package iofactory

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBufferedReaderTypical(t *testing.T) {
	testBufferedReader(t, 1<<16, 60000, 10)
}

func TestBufferedReaderLargeBufferSmallReadSize(t *testing.T) {
	testBufferedReader(t, 1<<20, 8000, 5)
}

func TestBufferedReaderSingleByte(t *testing.T) {
	testBufferedReader(t, 1, 1, 1024)
}

func TestBufferedReaderLarge(t *testing.T) {
	testBufferedReader(t, 1234567, 1000000, 1024)
}

func TestBufferedReaderLargeSmallBuffer(t *testing.T) {
	testBufferedReader(t, 12345, 1000, 100000)
}

func testBufferedReader(t *testing.T, length, readSize, iterations int) {

	buf := RandomByteSlice(length * iterations) // Require recycling buffers a few times
	bufCopied := make([]byte, len(buf)+100)
	assert.NotNil(t, buf)

	startTime := time.Now()
	randReader := NewMaxSizeReader(NewRandomSizeReader(bytes.NewReader(buf)), readSize)
	bufferedReader, err := NewBufferedReader(randReader, length, readSize)
	assert.Nil(t, err)

	n, err := io.ReadFull(bufferedReader, bufCopied)
	endTime := time.Now()

	MB := float64(len(buf)) / (1024 * 1024)
	elapsed := endTime.Sub(startTime)
	MBPerSecond := MB * 1000000000 / float64(elapsed.Nanoseconds())
	t.Logf("Length %v, Iterations %v -> %v MB, time %v, MB/s %0.3f", length, iterations, MB, elapsed, MBPerSecond)
	assert.True(t, strings.Contains(err.Error(), "EOF"))

	// Compare the buffers
	assert.True(t, len(buf) < len(bufCopied))
	assert.Equal(t, len(buf), length*iterations)
	assert.Equal(t, n, len(buf))
	assert.True(t, bytes.Equal(buf, bufCopied[:len(buf)]))
}

func BenchmarkBufferedReaderTypical(b *testing.B) {
	benchmarkBufferedReader(b, 1<<16, 60000, 10)
}

func BenchmarkBufferedReaderLargeBufferSmallReadSize(b *testing.B) {
	benchmarkBufferedReader(b, 1<<20, 8000, 5)
}

func BenchmarkBufferedReaderSingleByte(b *testing.B) {
	benchmarkBufferedReader(b, 1, 1, 1024)
}

func BenchmarkBufferedReaderLarge(b *testing.B) {
	benchmarkBufferedReader(b, 1<<24, 1000000, 8)
}

func benchmarkBufferedReader(b *testing.B, length, readSize, iterations int) {

	buf := RandomByteSlice(length * iterations) // Require recycling buffers a few times
	bufCopied := make([]byte, len(buf))

	b.Logf("Length %v, Iterations %v -> %v MB", length, iterations, float64(len(buf))/(1024*1024))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reader := NewMaxSizeReader(bytes.NewReader(buf), readSize)
		bufferedReader, _ := NewBufferedReader(reader, length, readSize)
		n, err := io.ReadFull(bufferedReader, bufCopied)
		panicIf(n < len(buf) || err != nil)
	}
}
