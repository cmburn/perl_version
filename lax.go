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

// This file holds the implementation of the parser for perl's lax versioning
// spec. It's largely deprecated by the Perl project, but there's still a good
// deal of modules on CPAN that use it.

import (
	"regexp"
	"strings"
)

var (
	laxRegexp = regexp.MustCompile(LaxVersionRegex)
)

type laxDotted struct {
	integer           string // version A
	dottedGroup       string
	alpha             string
	secondInteger     string // version B
	secondDottedGroup string
	secondAlpha       string
}

type laxDecimal struct {
	integer        string // version A
	fraction       string
	alpha          string
	secondFraction string // version B
	secondAlpha    string
}

type lax struct {
	original       string
	undef          string
	dotted         string
	dottedMatches  laxDotted
	decimal        string
	decimalMatches laxDecimal
}

func (d laxDotted) toPerlVersionA(original string) Version {
	dotted := d.dottedGroup
	isAlpha := d.alpha != ""
	if isAlpha {
		dotted += strings.TrimPrefix(d.alpha, "_")
	}
	var minors []int64
	if dotted != "" {
		minors = dottedToMinors(dotted)
	}
	numValues := len(minors)
	if numValues < 3 {
		// implied zeroes in v-qualified lax version
		numValues = 3
	}
	values := make([]int64, numValues)
	values[0] = mustParseInt64(d.integer)
	if minors != nil {
		copy(values[1:], minors)
	}
	return Version{
		original: original,
		alpha:    isAlpha,
		qv:       true,
		version:  values,
	}
}

func (d laxDotted) toPerlVersionB(original string) Version {
	// This particular case is a bit tricky. If there's three values,
	// *implied* zeroes included, it counts as a quoted lax version.

	dotted := d.secondDottedGroup
	isAlpha := d.secondAlpha != ""
	if isAlpha {
		dotted += strings.TrimPrefix(d.secondAlpha, "_")
	}
	minors := dottedToMinors(dotted)
	numValues := len(minors)
	impliedZero := d.secondDottedGroup[0] == '.' && d.secondInteger == ""
	if d.secondInteger != "" || impliedZero {
		numValues++
	}
	values := make([]int64, numValues)
	if d.secondInteger != "" {
		values[0] = mustParseInt64(d.secondInteger)
	} else if impliedZero {
		values[0] = 0
	}
	if minors != nil {
		copy(values[1:], minors)
	}

	return Version{
		original: original,
		alpha:    d.secondAlpha != "",
		qv:       numValues == 3,
		version:  values,
	}
}

func (d laxDotted) toPerlVersion(original string) Version {
	if d.integer != "" {
		return d.toPerlVersionA(original)
	} else if d.secondDottedGroup != "" {
		return d.toPerlVersionB(original)
	} else {
		panic("unreachable")
	}
}

func (d laxDecimal) toPerlVersionA(original string) (Version, error) {
	// Due to, what I can tell, is a runtime check, this is the only
	// subset of versioning that can error out. It happens when there's an
	// alpha part but no fractional part. It would make sense to change the
	// regex, but I'm hesitant to deviate from the Perl versioning spec.
	// Example: "1_0"

	fractionStr := d.fraction
	isAlpha := d.alpha != ""
	if isAlpha {
		if d.fraction == "" {
			return Version{}, errAlphaWithoutDecimal
		}
		fractionStr += strings.TrimPrefix(d.alpha, "_")
	}
	fractions := getFractionValue(fractionStr)
	numValues := len(fractions) + 1
	impliedZeroEnd := original[len(original)-1] == '.' && d.fraction == ""
	if impliedZeroEnd {
		numValues++
	}
	values := make([]int64, numValues)
	values[0] = mustParseInt64(d.integer)
	if fractions != nil {
		copy(values[1:], fractions)
	}
	if impliedZeroEnd {
		values[len(values)-1] = 0
	}
	return Version{
		original: original,
		alpha:    d.alpha != "",
		qv:       false,
		version:  values,
	}, nil
}

func (d laxDecimal) toPerlVersionB(original string) Version {
	fractionStr := d.secondFraction
	isAlpha := d.secondAlpha != ""
	if isAlpha {
		fractionStr += strings.TrimPrefix(d.secondAlpha, "_")
	}
	fractions := getFractionValue(fractionStr)
	if fractions == nil {
		panic("unreachable")
	}
	values := make([]int64, len(fractions)+1) // implied zero
	values[0] = 0
	copy(values[1:], fractions)
	return Version{
		original: original,
		alpha:    d.secondAlpha != "",
		qv:       false,
		version:  values,
	}
}

func (d laxDecimal) toPerlVersion(original string) (Version, error) {
	if d.integer != "" {
		return d.toPerlVersionA(original)
	} else if d.secondFraction != "" {
		return d.toPerlVersionB(original), nil
	} else {
		panic("unreachable")
	}
}

func (d lax) toPerlVersion() (Version, error) {
	if d.undef != "" {
		return Version{
			original: d.original,
			alpha:    false,
			qv:       false,
			version:  []int64{0},
		}, nil
	} else if d.dotted != "" {
		return d.dottedMatches.toPerlVersion(d.original), nil
	} else if d.decimal != "" {
		return d.decimalMatches.toPerlVersion(d.original)
	} else {
		panic("unreachable")
	}
}

func laxVersion(matches []string) (Version, error) {
	return lax{
		original: matches[0],
		undef:    matches[1],
		dotted:   matches[2],
		dottedMatches: laxDotted{
			integer:           matches[3],
			dottedGroup:       matches[4],
			alpha:             matches[5],
			secondInteger:     matches[6],
			secondDottedGroup: matches[7],
			secondAlpha:       matches[8],
		},
		decimal: matches[9],
		decimalMatches: laxDecimal{
			integer:        matches[10],
			fraction:       matches[11],
			alpha:          matches[12],
			secondFraction: matches[13],
			secondAlpha:    matches[14],
		},
	}.toPerlVersion()
}
