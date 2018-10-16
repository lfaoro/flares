package svc

import (
	"strings"
	"unicode"
)

// length of string
// extract positions
// replace at positions

// Mask replaces every other char of a string with an *.
func Mask(data string) string {
	count := -1
	state := true
	transform := func(r rune) rune {
		if unicode.IsPunct(r) {
			return r
		}
		if count < 3 {
			count++
		} else {
			state = !state
			count = 0
		}
		if state {
			return '*'
		}
		return r
	}
	return strings.Map(transform, data)
}
