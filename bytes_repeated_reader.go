package iofactory

import (
	"io"
)

type repeatedReader struct {
	data                  []byte
	iterationsTotal       int
	iterationsCompleted   int
	currentIterationIndex int
	loopForever           bool
}

// NewBytesRepeatedReader returns a reader that is similar to bytes.NewReader
// but allows N iterations.  If N is negative then this will repeat forever.
func NewBytesRepeatedReader(b []byte, iterations int) io.Reader {
	r := repeatedReader{
		data:            b,
		iterationsTotal: iterations,
		loopForever:     (iterations < 0),
	}
	return &r
}

func (r *repeatedReader) Read(p []byte) (totalWritten int, err error) {

	for len(p) > 0 && (r.loopForever || r.iterationsCompleted < r.iterationsTotal) {
		n := copy(p, r.data[r.currentIterationIndex:])
		p = p[n:]
		r.currentIterationIndex += n
		totalWritten += n

		if r.currentIterationIndex >= len(r.data) {
			r.currentIterationIndex = 0
			r.iterationsCompleted++
		}
	}

	if totalWritten == 0 && !r.loopForever && r.iterationsCompleted >= r.iterationsTotal {
		err = io.EOF
	}
	return
}
