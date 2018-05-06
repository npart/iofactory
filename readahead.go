package iofactory

import (
	"fmt"
	"io"
	"sync"
)

type buffer struct {
	data []byte
	id   int
}
type pbuffer *buffer

type readaheadChunk struct {
	data        []byte
	errToReturn error
	bufferID    int
}

// Readahead is
type Readahead struct {
	upstream        io.Reader
	leaseCount      []int
	chunksFilled    chan *readaheadChunk
	chunksToRecycle chan *readaheadChunk
	chunksEmpty     chan *readaheadChunk
	chunksAllocated []readaheadChunk
	buffers         []pbuffer
	freeBuffers     chan pbuffer
	exitSignal      chan bool
	bytesRead       int64
	currentChunk    *readaheadChunk
	closeOnce       sync.Once
}

// NewReadahead is used to create a readahead.
func NewReadahead(upstream io.Reader, bufferCount, bufferSize, readSize int) (*Readahead, error) {

	if bufferCount < 1 || bufferSize < 1 || bufferSize < readSize {
		return nil, fmt.Errorf("bufferCount must be at least 1, bufferSize must be positive, and bufferSize must be at least as large as readSize")
	}

	chunksToAllocate := maxInt(16, minInt(1024, 4*bufferCount*bufferSize/readSize))

	ra := Readahead{
		upstream:        upstream,
		leaseCount:      make([]int, bufferCount),
		chunksFilled:    make(chan *readaheadChunk, chunksToAllocate),
		chunksToRecycle: make(chan *readaheadChunk, chunksToAllocate),
		chunksEmpty:     make(chan *readaheadChunk, chunksToAllocate),
		chunksAllocated: make([]readaheadChunk, chunksToAllocate),
		buffers:         make([]pbuffer, bufferCount),
		freeBuffers:     make(chan pbuffer, bufferCount),
		exitSignal:      make(chan bool, 0),
	}

	for i := 0; i < bufferCount; i++ {
		newBuffer := buffer{
			id:   i,
			data: make([]byte, bufferSize),
		}
		ra.buffers[i] = &newBuffer
		ra.freeBuffers <- &newBuffer
	}

	for i := 0; i < chunksToAllocate; i++ {
		ra.chunksEmpty <- &ra.chunksAllocated[i]
	}

	// Run forever ...
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c := readaheadChunk{
					errToReturn: fmt.Errorf("readahead panic while reading: %v", r),
				}
				ra.chunksFilled <- &c
			}
		}()

		// Decrement the counter on the buffer
		decrementBufferLease := func(bufferID int) {
			newLeaseCount := ra.leaseCount[bufferID] - 1
			ra.leaseCount[bufferID] = newLeaseCount

			assertTrue(newLeaseCount >= 0)

			if newLeaseCount == 0 {
				ra.freeBuffers <- ra.buffers[bufferID]
			}
		}

		incrementBufferLease := func(bufferID int) {
			ra.leaseCount[bufferID]++
		}

		// Read from the upstream socket.
		recycleChunk := func(chunkToRecycle *readaheadChunk) {
			bufferID := chunkToRecycle.bufferID
			ra.chunksEmpty <- chunkToRecycle
			decrementBufferLease(bufferID)
		}

		for {
			// Get a buffer ...
			var nextBuffer pbuffer

			select {
			case chunkToRecycle := <-ra.chunksToRecycle:
				recycleChunk(chunkToRecycle)
				continue
			case nextBuffer = <-ra.freeBuffers:
				break
			case <-ra.exitSignal:
				return
			}

			panicIf(nextBuffer == nil)
			incrementBufferLease(nextBuffer.id)

			// Read and fill up this buffer one chunk at a time
			readOffset := 0
			for readOffset+readSize <= bufferSize {

				// Get a chunk to fill
				var chunkToFill *readaheadChunk
				select {
				case chunkToFill = <-ra.chunksEmpty:
					panicIf(chunkToFill == nil)
					break
				case chunkToRecycle := <-ra.chunksToRecycle:
					recycleChunk(chunkToRecycle)
					continue
				case <-ra.exitSignal:
					return
				}

				n, err := ra.upstream.Read(nextBuffer.data[readOffset : readOffset+readSize])

				if n > 0 {
					chunkToFill.data = nextBuffer.data[readOffset : readOffset+n]
					chunkToFill.bufferID = nextBuffer.id
					chunkToFill.errToReturn = nil

					incrementBufferLease(nextBuffer.id)
					ra.chunksFilled <- chunkToFill
					readOffset += n
				}

				if err != nil {
					// Use a newly allocated chunk for this to
					// avoid a potential deadlock.
					c := readaheadChunk{
						errToReturn: err,
					}
					ra.chunksFilled <- &c
					return
				}
			}

			decrementBufferLease(nextBuffer.id)
		}
	}()

	return &ra, nil
}

// Close will close the upstream reader.
func (ra *Readahead) Close() error {
	close(ra.exitSignal)

	var err error
	ra.closeOnce.Do(func() {
		closer, ok := ra.upstream.(io.Closer)
		if ok {
			err = closer.Close()
		}
	})

	return err
}

func (ra *Readahead) Read(p []byte) (totalCopied int, errToReturn error) {

	// Fill up the output as much as possible.
	// Quit when the provided buffer is full or we
	// don't have any more data.
	for len(p) > 0 {

		if ra.currentChunk == nil {
			select {
			case nextChunk := <-ra.chunksFilled:
				ra.currentChunk = nextChunk
			default:
				// No data available.  If we copied some data already, then return.
				// Else, if we haven't done anything yet, then we need to wait for
				// some data.
				if totalCopied > 0 {
					return
				}
				ra.currentChunk = <-ra.chunksFilled
			}
		}

		if len(ra.currentChunk.data) > 0 {
			n := copy(p, ra.currentChunk.data)
			totalCopied += n
			ra.currentChunk.data = ra.currentChunk.data[n:]
			p = p[n:]
		} else {
			if ra.currentChunk.errToReturn != nil {
				if totalCopied == 0 {
					// No data was copied, so set the err flag and return
					errToReturn = ra.currentChunk.errToReturn
				}
				return
			}

			// We have exhausted this chunk.  We can recycle this buffer.
			ra.chunksToRecycle <- ra.currentChunk
			ra.currentChunk = nil
		}
	}
	return
}
