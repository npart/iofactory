package iofactory

import (
	"io"
)

// NewRandomSizeReader returns a io.Reader, which is used as a passthrough
// reader that will read and return slices with random sizes.  It helps during testing
// to be intentionally inconsistent in read sizes to test for possible bugs / issues
// when reads are too consistent in time or size.
func NewRandomSizeReader(upstream io.Reader) io.Reader {
	panicIf(upstream == nil)

	r := randomSizeReader{
		upstream: upstream,
	}

	return &r
}

type randomSizeReader struct {
	upstream io.Reader
}

func (r *randomSizeReader) Read(p []byte) (int, error) {

	if len(p) == 0 {
		return r.upstream.Read(p)
	}

	sizeToRead := RandomInt(1, len(p), true)
	assertTrue(sizeToRead > 0 && sizeToRead <= len(p))
	return r.upstream.Read(p[:sizeToRead])
}
