package textparser

import (
    "unicode"
)

// This function is the default value for the `IsEscapeRune` field in
// `TokenScanner`. It returns true if the character is a reverse solidus (\).
func IsEscapeRune(ch rune) (bool) {
    if ch == '\\' {
        return true
    }

    return false
}

// This function is the default value for the `IsSymbolRune` field in
// `TokenScanner`.
func IsSymbolRune(ch rune) bool {
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

// This function is the default value for the `IsIdentRune` field in
// `TokenScanner`.
func IsIdentRune(ch rune, i int) bool {
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
//     "" - U+0022,U+0022
//     '' - U+0027,U+0027
//     “” - U+201C,U+201D
//     ‘’ - U+2018,U+2019
//     ‹› - U+2039,U+203A
//     «» - U+00AB,U+00BB
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
