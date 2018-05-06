// Copyright (c) 2018 Isaac Gremmer, released under MIT License. See LICENSE file.
package iofactory

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesLoopReaderHelloWorld(t *testing.T) {
	buf := []byte("hello world")
	length := len(buf)
	assert.True(t, length > 0)
	iterations := 3
	reader := NewBytesRepeatedReader(buf, iterations)

	output := make([]byte, length*(iterations+1)) // with extra space
	n, err := io.ReadAtLeast(reader, output, length*iterations)

	assert.Nil(t, err)
	assert.Equal(t, n, length*iterations)

	for i := 0; i < n; i++ {
		assert.Equal(t, buf[i%length], output[i])
	}
}

func TestBytesRepeatedReader(t *testing.T) {
	bufLen := 100
	repeats := 8000
	char := byte(' ')
	buf := make([]byte, bufLen)
	memset(buf, char)

	reader := NewMaxSizeReader(NewRandomSizeReader(NewBytesRepeatedReader(buf, repeats)), 500)
	output, _ := ioutil.ReadAll(reader)

	assert.Equal(t, len(output), bufLen*repeats)

	for index, _ := range output {
		if output[index] != char {
			t.Errorf("Byte index %v: expected %v, actual %v", index, int(char), int(output[index]))
		}
	}
}
