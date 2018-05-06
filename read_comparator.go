package iofactory

import (
	"io"
)

// CompareReaders will read the entire contents of the readers
// and compare that they are equal.  For simplicity this function
// will only read one byte at a time from each of the readers,
// which may not be performant enough for production but is
// plenty useful for unit testing.
func CompareReaders(firstReader, secondReader io.Reader, otherReaders ...io.Reader) bool {
	readers := append([]io.Reader{secondReader}, otherReaders...)

	buf := make([]byte, 1)
	buf2 := make([]byte, 1)

	for {
		n, err := firstReader.Read(buf)

		for _, reader := range readers {
			n2, err2 := reader.Read(buf2)

			if (n != n2) ||
				(n > 0 && buf[0] != buf2[0]) ||
				((err == nil) != (err2 == nil)) {
				return false
			}
		}

		if err != nil {
			break
		}
	}

	return true
}
