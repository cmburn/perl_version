// Copyright (c) 2022 Charlie Burnett
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package perl_version

import (
	"errors"
	"strconv"
	"strings"
)

var (
	errAlphaWithoutDecimal = errors.New("invalid version format: alpha " +
		"without decimal")
)

func mustParseInt64(s string) int64 {
	val, err := strconv.Atoi(s)
	if err != nil {
		// if we make it this far and fail, something's fundamentally
		// wrong with the code.
		panic(err)
	}
	return int64(val)
}

func dottedToMinors(s string) []int64 {
	s = strings.TrimPrefix(s, ".")
	raw := strings.Split(s, ".")
	minors := make([]int64, len(raw))
	for i, s := range raw {
		minors[i] = mustParseInt64(s)
	}
	return minors
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func getFractionValue(s string) []int64 {
	if s == "" {
		// should only happen in lax decimal shenanigans
		return nil
	}
	s = strings.TrimPrefix(s, ".")
	expectedValues := (len(s) / 3) + 1
	if (len(s) % 3) == 0 {
		expectedValues--
	}
	stringValues := make([]string, expectedValues)
	currentString := ""
	for i := range s {
		if i%3 == 0 && i != 0 {
			stringValues[i/3-1] = currentString
			currentString = ""
		}
		currentString += string(s[i])
	}
	// have to pad the last value with zeros until three wide
	for i := len(currentString); i < 3; i++ {
		currentString += "0"
	}
	stringValues[len(stringValues)-1] = currentString
	values := make([]int64, len(stringValues))
	for i, s := range stringValues {
		values[i] = mustParseInt64(s)
	}

	return values
}
