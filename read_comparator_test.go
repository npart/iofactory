package iofactory

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadComparatorEqual3(t *testing.T) {
	r1 := bytes.NewReader([]byte("hello world"))
	r2 := bytes.NewReader([]byte("hello world"))
	r3 := bytes.NewReader([]byte("hello world"))

	assert.True(t, CompareReaders(r1, r2, r3))
}

func TestReadComparatorNotEqualSize3(t *testing.T) {
	r1 := bytes.NewReader([]byte("hello world"))
	r2 := bytes.NewReader([]byte("hello world"))
	r3 := bytes.NewReader([]byte("hello"))

	assert.False(t, CompareReaders(r1, r2, r3))
}

func TestReadComparatorNotEqualSize2(t *testing.T) {
	r1 := bytes.NewReader([]byte("hello"))
	r2 := bytes.NewReader([]byte("hello world"))

	assert.False(t, CompareReaders(r1, r2))
}

func TestReadComparatorNotEqualContent3(t *testing.T) {
	r1 := bytes.NewReader([]byte("hello world"))
	r2 := bytes.NewReader([]byte("hello vqrld"))
	r3 := bytes.NewReader([]byte("hello world"))

	assert.False(t, CompareReaders(r1, r2, r3))
}

func TestReadComparatorNotEqualContent2(t *testing.T) {
	r1 := bytes.NewReader([]byte("hello world"))
	r2 := bytes.NewReader([]byte("hello vqrld"))

	assert.False(t, CompareReaders(r1, r2))
}
