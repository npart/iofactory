// Copyright (c) 2018 Isaac Gremmer, released under MIT License. See LICENSE file.
package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/npart/iofactory"
)

// These are the examples that are part of the README.md

func exampleMaxSizeReader() {
	// Read from file
	input, _ := os.Open("input.txt")
	chunkedInput := iofactory.NewMaxSizeReader(input, 1024) // read the whole file, but only read 1024 bytes at a time
	io.Copy(os.Stdout, chunkedInput)

}
func exampleMaxSizeSocketServer() {
	// Network (echo back to client) (requires client to be functional)
	ln, _ := net.Listen("tcp", ":8080")
	for {
		conn, _ := ln.Accept()
		go func() {
			defer conn.Close()
			chunkedInput := iofactory.NewMaxSizeReader(conn, 16) // echo full input stream, but a maximum of 16 bytes at a time
			io.Copy(conn, chunkedInput)
		}()
	}
}

func exampleRandomSizeReader() {
	// File
	input, _ := os.Open("input.txt")
	defer input.Close()
	randomSizeReader := iofactory.NewRandomSizeReader(input)
	totalRead := 0
	for {
		buf := make([]byte, 1024)
		n, err := randomSizeReader.Read(buf) // each read will be between 1 and 1024 bytes
		totalRead += n
		if err != nil {
			break
		}
	}
	log.Printf("Read %v bytes from file", totalRead)
}

func exampleRandomAndMaxSizeReader() {
	// Read with maximum and random read size
	input, _ := os.Open("input.txt")
	defer input.Close()
	randomSizeReader := iofactory.NewMaxSizeReader(iofactory.NewRandomSizeReader(input), 128)

	// the whole file will be read to buf, but each read will be between 1 and 128 bytes
	// even though ioutil.ReadAll() will call Read() with a much larger buffer size
	buf, _ := ioutil.ReadAll(randomSizeReader)
	log.Printf("%v", string(buf))
}

func exampleBytesRepeatedReader() {
	buf := []byte("hello world")
	reader := iofactory.NewBytesRepeatedReader(buf, 3)
	io.Copy(os.Stdout, reader)
}

func exampleBufferedReader() {
	buf := iofactory.RandomByteSlice(1 << 24) // 16 MB of random data
	reader := bytes.NewReader(buf)
	bufferedReader, err := iofactory.NewBufferedReader(reader, 1<<20, 1<<16) // Buffer size of 1MB, read 64KB
	if err != nil {
		log.Printf("err %v", err)
		return
	}
	io.Copy(os.Stdout, bufferedReader)
}

func main() {
	exampleMaxSizeReader()
	exampleRandomSizeReader()
	exampleRandomAndMaxSizeReader()
	exampleBytesRepeatedReader()
}
