// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package assert

import (
	"regexp"
	"slices"
	"strings"
)

// EndsNotWith asserts that the string `str` does NOT end with the given rune `suffix`.
// If it does, the function panics with the provided `msg`.
func EndsNotWith(str string, suffix rune, msg string) {
	if len(str) > 0 && str[len(str)-1] == byte(suffix) {
		panic(msg)
	}
}

// EndsNotWith asserts that the string `str` does NOT end with the given rune `suffix`.
// If it does, the function panics with the provided `msg`.
func EndsWith(str string, suffix string, msg string) {
	if !strings.HasSuffix(str, suffix) {
		panic(msg)
	}
}

// IsAsciiString asserts that the string `s` contains only ASCII characters (code <= 127).
// If a non-ASCII character is found, the function panics with the provided `msg`.
func IsAsciiString(s string, msg string) {
	for i := range len(s) {
		if s[i] > 127 {
			panic(msg)
		}
	}
}

// GreaterThanOrEqual asserts that integer `i` is greater than or equal to `lower`.
// If `i < lower`, the function panics with the provided `msg`.
func GreaterThanOrEqual(i int, lower int, msg string) {
	if i < lower {
		panic(msg)
	}
}

// StartsNotWith asserts that the string `str` does NOT start with the given rune `prefix`.
// If it does, the function panics with the provided `msg`.
func StartsNotWith(str string, prefix rune, msg string) {
	if len(str) > 0 && str[0] == byte(prefix) {
		panic(msg)
	}
}

// StartsNotWith asserts that the string `str` does NOT start with the given `prefix`.
// If it does, the function panics with the provided `msg`.
func StartsNotWithString(str string, prefix string, msg string) {
	if strings.HasPrefix(str, prefix) {
		panic(msg)
	}
}

// StartsWith asserts that the string `str` does start with the given `prefix`.
// If it does, the function panics with the provided `msg`.
func StartsWith(str string, prefix string, msg string) {
	if !strings.HasPrefix(str, prefix) {
		panic(msg)
	}
}

// Unreachable should be called at code paths that must never be reached.
// It always panics with the provided `msg`.
func Unreachable(msg string) {
	panic(msg)
}

// True asserts that the boolean `b` is true.
// If not, it panics with the provided `msg`.
func True(b bool, msg string) {
	if !b {
		panic(msg)
	}
}

// False asserts that the boolean `b` is false.
// If not, it panics with the provided `msg`.
func False(b bool, msg string) {
	if b {
		panic(msg)
	}
}

// IsAlphaDash asserts that the string `s` contains only letters (A-Z, a-z) and dashes (-).
// If not, it panics with the provided `msg`.
func IsAlphaDash(s string, msg string) {
	if matched, _ := regexp.MatchString(`^[A-Za-z-]+$`, s); !matched {
		panic(msg)
	}
}

// ErrNil asserts that the error `err` is nil.
// If not, it panics with the error.
func ErrNil(err error) {
	if err != nil {
		panic(err)
	}
}

// NotNil asserts that the object `o` is not nil.
// If it is nil, it panics with the provided `msg`.
func NotNil(o any, msg string) {
	if o == nil {
		panic(msg)
	}
}

func In[T comparable](element T, arr []T, msg string) {
	if slices.Contains(arr, element) {
		return
	}
	panic(msg)
}
