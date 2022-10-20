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

// Package perl_version is a Go implementation of Perl's version.pm. It's
// written for the purpose of working with Perl packages in the context of a
// larger, multi-language monorepo. It's written to be a bug-for-bug compatible
// implementation of Perl's version.pm.
package perl_version

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Version is a direct, mapping of Perl's version::Internal, methods and
// all. It's meant to be opaque, as the internal representation might change
// if the need arises.
type Version struct {
	original string
	alpha    bool
	qv       bool
	version  []int64
}

///////////////////////////////////////////////////////////////////////////////
// Direct methods from Perl                                                  //
///////////////////////////////////////////////////////////////////////////////

// These methods are directly copied from Perl's version.pm.

// IsAlpha checks whether a version is an alpha version. This is implied by the
// presence of an underscore in the version. For example, "1.2_3" is an alpha
// version (equating to "v1.203.0", but that's due to Perl's bizarre version
// semantics).
func (v *Version) IsAlpha() bool {
	return v.alpha
}

// IsQv checks whether a version is a qv version. This is indicated by a 'v' at
// the beginning of the version. For example, "v1.2.3" is a qv version, while
// "1.2.3" is not. The versions are not equal either- "1.2.3" is represented
// as "v1.200.300", and "v1.2.3" is still just "v1.2.3".
func (v *Version) IsQv() bool {
	return v.qv
}

// Normal is a convenience function for normalizing a version string. It
// returns it in standardized qv form, with at least three subversions.
func (v *Version) Normal() string {
	num := len(v.version)
	if num < 3 {
		num = 3
	}
	fixed := make([]int64, num)
	copy(fixed, v.version)
	asStrings := make([]string, num)
	for i, v := range fixed {
		asStrings[i] = fmt.Sprintf("%d", v)
	}
	return "v" + strings.Join(asStrings, ".")
}

// Numify returns the numeric version of a version string. For example,
// "v1.2.3" would return 1.002003. This is useful for quick comparisons, and
// embedding in maps, though if you have a version with many subversions, it's
// probably better to use the relevant comparison methods (which are probably
// faster regardless).
func (v *Version) Numify() float64 {
	if len(v.version) == 1 {
		return float64(v.version[0])
	}
	asStrings := make([]string, len(v.version)-1)
	for i, v := range v.version[1:] {
		asStrings[i] = strconv.FormatInt(v, 10)
		// pad with zeros
		for len(asStrings[i]) < 3 {
			asStrings[i] = "0" + asStrings[i]
		}
	}
	tail := strings.Join(asStrings, "")
	str := fmt.Sprintf("%d.%s", v.version[0], tail)
	out, _ := strconv.ParseFloat(str, 64)
	return out
}

// Stringify matches its Perl equivalent- functionally it acts the same as Raw,
// however if the Version is undefined, it returns "0".
func (v *Version) Stringify() string {
	if v.original == "undef" {
		return "0"
	}
	return v.original
}

// Raw returns the original representation of the version.
func (v *Version) Raw() string {
	return v.original
}

// MarshalJSON implements the json.Marshaler interface. This allows for caching
// of the version.
func (v *Version) MarshalJSON() ([]byte, error) {
	data := struct {
		Original string  `json:"original"`
		Alpha    bool    `json:"alpha"`
		Qv       bool    `json:"qv"`
		Version  []int64 `json:"version"`
	}{
		Original: v.original,
		Alpha:    v.alpha,
		Qv:       v.qv,
		Version:  v.version,
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements the json.Unmarshaler interface. This allows for
// extracting the version from a cached version.
func (v *Version) UnmarshalJSON(data []byte) error {
	var obj struct {
		Original string  `json:"original"`
		Alpha    bool    `json:"alpha"`
		Qv       bool    `json:"qv"`
		Version  []int64 `json:"version"`
	}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return err
	}
	v.original = obj.Original
	v.alpha = obj.Alpha
	v.qv = obj.Qv
	v.version = obj.Version
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// Comparisons                                                               //
///////////////////////////////////////////////////////////////////////////////

// LessThan checks whether a version is older than another.
func (v *Version) LessThan(other *Version) bool {
	length := min(len(v.version), len(other.version))
	for i := 0; i < length; i++ {
		if v.version[i] < other.version[i] {
			return true
		}
		if v.version[i] > other.version[i] {
			return false
		}
	}
	return false
}

// GreaterThan checks whether a version is newer than another.
func (v *Version) GreaterThan(other *Version) bool {
	length := min(len(v.version), len(other.version))
	for i := 0; i < length; i++ {
		if v.version[i] > other.version[i] {
			return true
		}
		if v.version[i] < other.version[i] {
			return false
		}
	}
	return false
}

// Equal checks whether two versions are the same. This doesn't strictly
// mean they're identical, it means, for example, "v5.34" counts as the same as
// "v5.34.0" *or* "v5.34.1".
func (v *Version) Equal(other *Version) bool {
	return !(v.LessThan(other) || v.GreaterThan(other))
}

// LessThanOrEqual checks whether a version is older or equivalent to
// another. Same as (LessThan || Equal).
func (v *Version) LessThanOrEqual(other *Version) bool {
	return v.Equal(other) || v.LessThan(other)
}

// GreaterThanOrEqual checks whether a version is newer or equivalent to
// another. Same as (GreaterThan || Equal).
func (v *Version) GreaterThanOrEqual(other *Version) bool {
	return v.Equal(other) || v.GreaterThan(other)
}

// NotEqual checks whether two versions are not the same. Same as
// !(Equal).
func (v *Version) NotEqual(other *Version) bool {
	return !v.Equal(other)
}

// Compare compares two versions. It returns -1 if the receiver is older,
// 0 if they're equivalent, and 1 if the receiver is newer.
func (v *Version) Compare(other *Version) int {
	if v.LessThan(other) {
		return -1
	}
	if v.GreaterThan(other) {
		return 1
	}
	return 0
}

func init() {
	strictRegexp.Longest()
	laxRegexp.Longest()
}
