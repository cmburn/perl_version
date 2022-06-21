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

// Keep the regex in a separate file to make it easier on the eyes.

// notation:
// suffix:
// R: regex constant
// P: repeating/plus group
// ${N}P: repeating N or more times
// Nc: Non capturing group

// shared between lax and strict:
const (
	fractionR = `(\.[0-9]+)`
)

// strict regexes:
const (
	strictIntR         = `(0|[1-9][0-9]*)`
	strictDottedNcR    = `(?:\.[0-9]{1,3})`
	strictDotted2PR    = `(` + strictDottedNcR + `{2,})`
	strictDecimalFormR = `(` + strictIntR + fractionR + `?)`
	strictDottedFormR  = `(v` + strictIntR + strictDotted2PR + `)`
)

// lax regexes:
const (
	laxIntR         = `([0-9]+)`
	laxDottedNcR    = `(?:\.[0-9]+)`
	laxDotted2PR    = `(` + laxDottedNcR + `{2,})`
	laxDottedPR     = `(` + laxDottedNcR + `+)`
	laxAlphaR       = `(_[0-9]+)`
	laxUndefR       = `(undef)`
	laxDecimalFormR = `(` + laxIntR + `(?:` + fractionR + `|\.)?` +
		laxAlphaR + `?|` + fractionR + laxAlphaR + `?)`
	laxDottedFormR = `(v` + laxIntR + `(?:` + laxDottedPR + laxAlphaR +
		`?)?|` + laxIntR + `?` + laxDotted2PR + laxAlphaR + `?)`
)

// LaxVersionRegex is a regular expression that matches a Perl version string,
// under the documented rules under version::regexp. It is a direct adaptation
// to Go's regex-engine. The Lax version has a few interesting edge cases, but
// so there's actually four different forms it has to cover.
const LaxVersionRegex = `(?:` + laxUndefR + `|` + laxDottedFormR + `|` +
	laxDecimalFormR + `)$`

// StrictVersionRegex is a regular expression that matches a Perl version
// string,under the documented rules under version::regexp. Strict versioning
// is highly recommended, both by the Perl project and someone who's just
// had to write a parser for the lax version.
const StrictVersionRegex = `(?:` + strictDecimalFormR + `|` +
	strictDottedFormR + `)$`
