// Copyright (c) 2018 Isaac Gremmer, released under MIT License. See LICENSE file.
package iofactory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinInt(t *testing.T) {

	assert.Equal(t, minInt(3, -1, 5, 2), -1)
	assert.Equal(t, minInt(100), 100)
	assert.Equal(t, minInt(33, 10), 10)
}

func TestMaxInt(t *testing.T) {

	assert.Equal(t, maxInt(3, -1, 5, 2), 5)
	assert.Equal(t, maxInt(100), 100)
	assert.Equal(t, maxInt(33, 10), 33)
}

func TestRandInt(t *testing.T) {
	for _, minMaxVal := range [][]int{[]int{5, 10, 100}, []int{1, 2, 100}, []int{0, 5, 100}} {
		minVal, maxVal, iterations := minMaxVal[0], minMaxVal[1], minMaxVal[2]

		for _, inclusive := range []bool{false, true} {

			var upperExpectedVal int
			if inclusive {
				upperExpectedVal = maxVal
			} else {
				upperExpectedVal = maxVal - 1
			}

			hitCounts := make([]int, upperExpectedVal+1)

			for i := 0; i < iterations; i++ {
				v := RandomInt(minVal, maxVal, inclusive)
				if v >= minVal && v <= upperExpectedVal {
					hitCounts[v]++
				} else {
					t.Errorf("Value out of range: %v not between [%v, %v)", v, minVal, maxVal)
				}
			}

			for i := minVal; i <= upperExpectedVal; i++ {
				// Given enough iterations, each of these values would be selected at least once.
				assert.True(t, hitCounts[i] > 0)
			}
		}
	}
}

func TestRandomByteSlice(t *testing.T) {
	length := 1 << 24
	buf := RandomByteSlice(length)
	assert.Equal(t, len(buf), length)

	// Test each value of 0-255 is present and that we are not
	// experiencing long streaks of the same value.  Technically,
	// it's possible to get all of the same value as a random
	// sample, but it's very unlikely.
	var longestConsecutiveStreak int
	var currentConsecutiveStreak int
	var previousValue byte
	hitCounts := make([]int, 256)

	for index, b := range buf {
		hitCounts[b]++

		if index > 0 && b == previousValue {
			currentConsecutiveStreak++
			longestConsecutiveStreak = maxInt(longestConsecutiveStreak, currentConsecutiveStreak)
		} else {
			currentConsecutiveStreak = 1
			previousValue = b
		}
	}

	for _, hits := range hitCounts {
		assert.True(t, hits > 0)
	}

	assert.True(t, longestConsecutiveStreak >= 3)
	assert.True(t, longestConsecutiveStreak < 10)
}
