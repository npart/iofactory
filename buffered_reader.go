// Copyright (c) 2018 Isaac Gremmer, released under MIT License. See LICENSE file.
package iofactory

import (
	"fmt"
	"io"
	"sync"
)

// BufferedReader is
type BufferedReader struct {
	upstream                  io.Reader
	totalRead                 int64
	totalReadReturned         int64
	totalWritten              int64
	currentReadChunkAvailable int
	currentReadChunkSize      int
	buffer                    []byte
	bufferSize                int64
	readsAvailable            chan int
	readsCompleted            chan int
	exitSignal                chan bool
	closeOnce                 sync.Once
	errToReturn               error
}

// NewBufferedReader is used to create a new buffered reader (BufferedReader).
func NewBufferedReader(upstream io.Reader, bufferSize, readSize int) (*BufferedReader, error) {

	if bufferSize < 1 || bufferSize < readSize {
		return nil, fmt.Errorf("bufferSize must be positive, and bufferSize must be at least as large as readSize")
	}

	br := BufferedReader{
		upstream:       upstream,
		buffer:         make([]byte, bufferSize),
		bufferSize:     int64(bufferSize),
		readsAvailable: make(chan int, 1024),
		readsCompleted: make(chan int, 1024),
		exitSignal:     make(chan bool, 0),
	}

	// Run forever ...
	go func() {

		defer func() {
			close(br.readsAvailable)
		}()

		defer func() {
			if r := recover(); r != nil {
				br.errToReturn = fmt.Errorf("readahead panic while reading: %v", r)
			}
		}()

		spaceAvailable := func() int64 {
			return br.bufferSize - (br.totalWritten - br.totalReadReturned)
		}

		for {
			writePosition := br.totalWritten % br.bufferSize
			nextWriteSize := int64(minInt(int(br.bufferSize-writePosition), int(readSize)))

			for spaceAvailable() < nextWriteSize {
				select {
				case r := <-br.readsCompleted:
					if r == 0 {
						return
					}
					br.totalReadReturned += int64(r)
					continue
				case <-br.exitSignal:
					return
				}
			}

			n, err := br.upstream.Read(br.buffer[writePosition : writePosition+nextWriteSize])

			if n > 0 {
				br.totalWritten += int64(n)
				br.readsAvailable <- n
			}

			if err != nil {
				br.errToReturn = err
				return
			}

		}
	}()

	return &br, nil
}

// Close will close the upstream reader.
func (br *BufferedReader) Close() error {
	close(br.exitSignal)

	var err error
	br.closeOnce.Do(func() {
		closer, ok := br.upstream.(io.Closer)
		if ok {
			err = closer.Close()
		}
	})

	return err
}

func (br *BufferedReader) Read(p []byte) (totalCopied int, errToReturn error) {

	// Fill up the output as much as possible.
	// Quit when the provided buffer is full or we
	// don't have any more data.
	for len(p) > 0 {

		if br.currentReadChunkAvailable == 0 {
			select {
			case br.currentReadChunkSize = <-br.readsAvailable:
				if br.currentReadChunkSize == 0 {
					errToReturn = br.errToReturn
					return
				}
				br.currentReadChunkAvailable = br.currentReadChunkSize
			default:
				// No data available.  If we copied some data already, then return.
				// Else, if we haven't done anything yet, then we need to wait for
				// some data.
				if totalCopied > 0 {
					return
				}
				br.currentReadChunkSize = <-br.readsAvailable
				if br.currentReadChunkSize == 0 {
					errToReturn = br.errToReturn
					return
				}
				br.currentReadChunkAvailable = br.currentReadChunkSize
			}
		}

		panicIf(br.currentReadChunkAvailable == 0)
		readPosition := br.totalRead % br.bufferSize
		nextReadSize := int64(minInt(int(br.bufferSize-readPosition), int(br.currentReadChunkAvailable), len(p)))

		n := copy(p, br.buffer[readPosition:readPosition+nextReadSize])
		totalCopied += n
		br.totalRead += int64(n)
		br.currentReadChunkAvailable -= n
		p = p[n:]

		if br.currentReadChunkAvailable == 0 {
			br.readsCompleted <- br.currentReadChunkSize
		}
	}
	return
}
