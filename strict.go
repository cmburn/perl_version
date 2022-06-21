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
	"regexp"
	"strings"
)

var (
	strictRegexp = regexp.MustCompile(StrictVersionRegex)
)

type strictDecimalForm struct {
	integerPart  string
	fractionPart string
}

type strictDottedForm struct {
	integerPart string
	dottedGroup string
}

type strict struct {
	original       string
	decimal        string
	decimalMatches strictDecimalForm
	dotted         string
	dottedMatches  strictDottedForm
}

func (d strictDecimalForm) toPerlVersion(original string) Version {
	pv := Version{
		original: original,
		alpha:    false,
		qv:       false,
	}
	trimmed := strings.TrimPrefix(d.fractionPart, ".")
	fracValues := getFractionValue(trimmed)
	pv.version = make([]int64, len(fracValues)+1)
	pv.version[0] = mustParseInt64(d.integerPart)
	copy(pv.version[1:], fracValues)
	return pv
}

func (d strictDottedForm) toPerlVersion(original string) Version {
	pv := Version{
		original: original,
		alpha:    false,
		qv:       true,
	}
	trimmed := strings.TrimPrefix(d.dottedGroup, ".")
	minors := strings.Split(trimmed, ".")
	pv.version = make([]int64, len(minors)+1)
	pv.version[0] = mustParseInt64(d.integerPart)
	for i, part := range minors {
		pv.version[i+1] = mustParseInt64(part)
	}
	return pv
}

func (d strict) toPerlVersion() Version {
	if d.decimal != "" {
		return d.decimalMatches.toPerlVersion(d.original)
	} else if d.dotted != "" {
		return d.dottedMatches.toPerlVersion(
			d.original)
	} else {
		panic("logic error: strictRegexp matched but no version found")
	}
}

func strictVersion(matches []string) Version {
	return strict{
		original: matches[0],
		decimal:  matches[1],
		decimalMatches: strictDecimalForm{
			integerPart:  matches[2],
			fractionPart: matches[3],
		},
		dotted: matches[4],
		dottedMatches: strictDottedForm{
			integerPart: matches[5],
			dottedGroup: matches[6],
		},
	}.toPerlVersion()
}
