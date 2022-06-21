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
)

// Here are functions for working with strings as perl versions. Generally just
// wrappers around the respective methods on the Version type, so if you're
// comparing versions repeatedly, you should use the Version type directly.

// Parse parses a string into a Version. The string can be either a lax or
// strict versioning scheme, as defined in version::Internals.
func Parse(version string) (Version, error) {
	laxMatch := laxRegexp.FindStringSubmatch(version)
	strictMatch := strictRegexp.FindStringSubmatch(version)

	// lax needs to be checked first, since it can throw an error
	if laxMatch != nil {
		if strictMatch == nil {
			return laxVersion(laxMatch)
		}
		if len(laxMatch[0]) > len(strictMatch[0]) {
			lax, err := laxVersion(laxMatch)
			if err == nil {
				return lax, nil
			}
		}
	}

	// try strict next
	if strictMatch != nil {
		return strictVersion(strictMatch), nil
	}

	return Version{}, errors.New("invalid version string: " + version)
}

// Undef returns a new, undefined version.
func Undef() Version {
	return Version{
		original: "undef",
		alpha:    false,
		qv:       false,
		version:  []int64{0},
	}
}

func parseMulti(a, b string) (Version, Version, error) {
	aPv, err := Parse(a)
	if err != nil {
		return Version{}, Version{}, err
	}
	bPv, err := Parse(b)
	if err != nil {
		return Version{}, Version{}, err
	}
	return aPv, bPv, nil
}

// CompatibleWith checks if candidate is compatible with target. Panics if
// it can't parse either of the two; this is meant for already-validated
// versions. If you need to check, use the respective Version method
// Version.GreaterThanOrEqual.
func CompatibleWith(candidate, target string) bool {
	candidatePv, targetPv, err := parseMulti(candidate, target)
	if err != nil {
		panic(err)
	}
	return candidatePv.GreaterThanOrEqual(&targetPv)
}

// MustParse is for parsing a version string that must be valid. It panics
// if it can't parse the string. You probably want Parse(), unless you're
// dealing with an internal cache.
func MustParse(version string) Version {
	v, err := Parse(version)
	if err != nil {
		panic(err)
	}
	return v
}

// IsValid returns true if the version is parseable.
func IsValid(version string) bool {
	_, err := Parse(version)
	return err == nil
}
