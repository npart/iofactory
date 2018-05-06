package iofactory

import (
	"crypto/rand"
	"io"
	"math/big"
)

func assertTrue(b bool) {
	if !b {
		panic(b)
	}
}

func assertNil(i interface{}) {
	if i != nil {
		panic(i)
	}
}

func panicIf(b bool) {
	if b {
		panic(b)
	}
}

func minInt(firstNum int, nums ...int) int {
	minVal := firstNum

	for _, num := range nums {
		if num < minVal {
			minVal = num
		}
	}
	return minVal
}

func maxInt(firstNum int, nums ...int) int {
	maxVal := firstNum

	for _, num := range nums {
		if num > maxVal {
			maxVal = num
		}
	}
	return maxVal
}

func memset(p []byte, b byte) {
	for index := range p {
		p[index] = b
	}
}

// RandomInt will select a random integer between [minVal, maxVal) or
// [minVal, maxVal] depending on if 'inclusive' is set.
func RandomInt(minVal, maxVal int, inclusive bool) int {
	spread := maxVal - minVal
	if inclusive {
		spread++
	}
	randInt, err := rand.Int(rand.Reader, big.NewInt(int64(spread)))
	assertNil(err)
	return int(randInt.Int64()) + minVal
}

// RandomByteSlice will return a slice of random data of type []byte.
func RandomByteSlice(size int) []byte {
	buf := make([]byte, size)
	io.ReadFull(rand.Reader, buf)
	return buf
}
