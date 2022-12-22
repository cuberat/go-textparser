// BSD 2-Clause License
//
// Copyright (c) 2020 Don Owens <don@regexguy.com>.  All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice,
//   this list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package textparser

import (
	"unicode"
)

// This function is the default value for the `IsEscapeRune` field in
// `TokenScanner`. It returns true if the character is a reverse solidus (\).
// Where `i` is the index of `ch` in the current token parse, and `runes` is
// the list of runes already excepted for the current token.
func IsEscapeRune(ch rune, i int, runes []rune) bool {
	if ch == '\\' {
		return true
	}

	return false
}

// This function is the default value for the `IsSymbolRune` field in
// `TokenScanner`. Where `i` is the index of `ch` in the current token parse,
// and `runes` is the list of runes already excepted for the current token.
func IsSymbolRune(ch rune, i int, runes []rune) bool {
	if i > 0 {
		return false
	}
	if ok, _ := IsQuoteRune(ch); ok {
		return false
	}

	if unicode.IsSymbol(ch) {
		return true
	}

	if unicode.IsPunct(ch) {
		return true
	}

	return false
}

// This function is the default value for the `IsDigitRune` field in
// `TokenScanner`. Where `i` is the index of `ch` in the current token parse,
// and `runes` is the list of runes already excepted for the current token.
func IsDigitRune(ch rune, i int, runes []rune) bool {
	return unicode.IsDigit(ch)
}

// This function is the default value for the `IsIdentRune` field in
// `TokenScanner`. Where `i` is the index of `ch` in the current token parse,
// and `runes` is the list of runes already excepted for the current token.
func IsIdentRune(ch rune, i int, runes []rune) bool {
	if unicode.IsLetter(ch) {
		return true
	}

	if ch == '_' {
		return true
	}

	if unicode.IsPunct(ch) {
		return false
	}

	if i > 0 && unicode.IsDigit(ch) {
		return true
	}

	if unicode.IsMark(ch) {
		return true
	}

	return false
}

// Alternative predicate for determing opening quote characters. Set the
// `IsQuoteRune` field `TokenScanner` to this function to consider the
// following to be quote characters (the plain quotes from English, as well as
// more fancy ones specified in Unicode).
//
//	"" - U+0022,U+0022
//	'' - U+0027,U+0027
//	“” - U+201C,U+201D
//	‘’ - U+2018,U+2019
//	‹› - U+2039,U+203A
//	«» - U+00AB,U+00BB
func IsQuoteRuneFancy(ch rune) (bool, rune) {
	switch ch {
	case 0x0022: // ""
		return true, 0x0022
	case 0x0027: // ''
		return true, 0x0027
	case 0x0060: // ``
		return true, 0x0060
	case 0x201C: // “”
		return true, 0x201D
	case 0x2018: // ‘’
		return true, 0x2019
	case 0x2039: // ‹›
		return true, 0x203A
	case 0x00AB: // «»
		return true, 0x00BB
	}

	return false, 0
}

// This function is the default value for the `IsQuoteRune` field in
// `TokenScanner`. It only treats single quotes ('), double quotes ("), and
// back ticks (`) as quote characters.
func IsQuoteRune(ch rune) (bool, rune) {
	switch ch {
	case 0x0022: // ""
		return true, 0x0022
	case 0x0027: // ''
		return true, 0x0027
	case 0x0060: // ``
		return true, 0x0060
	}

	return false, 0
}

// This function is the default value for the `IsIdentRune` field in
// `TokenScanner`. Where `i` is the index of `ch` in the current token parse,
// and `runes` is the list of runes already excepted for the current token.
func IsSpaceRune(ch rune, i int, runes []rune) bool {
	if unicode.IsSpace(ch) {
		return true
	}

	return false
}
