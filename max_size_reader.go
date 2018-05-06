package iofactory

import (
	"io"
)

// NewMaxSizeReader returns a io.Reader, which is used as a passthrough
// reader that will read and return slices but with a maximum read size.
func NewMaxSizeReader(upstream io.Reader, maxReadSize int) io.Reader {
	panicIf(upstream == nil || maxReadSize < 1)

	r := maxSizeReader{
		upstream:    upstream,
		maxReadSize: maxReadSize,
	}

	return &r
}

type maxSizeReader struct {
	upstream                 io.Reader
	minReadSize, maxReadSize int
}

func (r *maxSizeReader) Read(p []byte) (int, error) {

	s := minInt(len(p), r.maxReadSize)
	return r.upstream.Read(p[:s])
}
