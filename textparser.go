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

// Package textparser provides a scanner and tokenizer for UTF-8-encoded text.
// It takes an io.Reader providing the source, which then can be tokenized
// through repeated calls to the Scan function. Use the NewScanner() function
// to create and initalize a scanner. In addition, convenience functions are
// provided to initialize a scanner from a byte slice or a string.
//
// This module was originally targeted at parsing non-trivial struct tags, but
// it was then made more general so that it could be used for other purposes.
//
// TokenScanner recognizes whitespace, comments (C++ style line and
// multi-line), symbols (e.g., used for to denote operators), string literals
// (quoted using one of " ' `), integers, and floats. By default, a
// TokenScanner skips white space and comments, identifiers begin with a
// letter and may contain letters or digits or "_" after the first character.
// It may be customized to recognize additional quotes or define identifiers
// differently, etc.
//
// Example:
//     src := `
//     // This is a comment.
//     if a > 5 {
//         b = "this is a string";
//         c = 7.2;
//     }
//     `
//     s := textparser.NewScanner(strings.NewReader(src))
//     s.SetFilename("nofile")
//
//     for s.Scan() {
//         if err := s.Err(); err != nil {
//             panic(fmt.Sprintf("error at %s: %s", s.Position(), err))
//         }
//         token := s.Token()
//         fmt.Printf("%-16s - %-6s -> %s\n", s.Position(), token.Type,
//             token.Text)
//     }
//
// Output:
//  nofile:3:5 (31)  - Ident  -> if
//  nofile:3:8 (34)  - Ident  -> a
//  nofile:3:10 (36) - Symbol -> >
//  nofile:3:12 (38) - Int    -> 5
//  nofile:3:14 (40) - Symbol -> {
//  nofile:4:9 (50)  - Ident  -> b
//  nofile:4:11 (52) - Symbol -> =
//  nofile:4:13 (54) - String -> "this is a string"
//  nofile:4:30 (72) - Symbol -> ;
//  nofile:5:9 (82)  - Ident  -> c
//  nofile:5:11 (84) - Symbol -> =
//  nofile:5:13 (86) - Float  -> 7.2
//  nofile:5:16 (89) - Symbol -> ;
//  nofile:6:5 (95)  - Symbol -> }
package textparser

import (
    "bufio"
    "bytes"
    "fmt"
    "io"
    "strings"
    "unicode"
    utf8 "unicode/utf8"
)

type TokenType int

// Supported token types.
const (
    TokenTypeWhitespace TokenType = iota
    TokenTypeIdent
    TokenTypeString
    TokenTypeComment
    TokenTypeInt
    TokenTypeFloat
    TokenTypeSymbol
)

// Returns a string representation of the token type.
func (t TokenType) String() string {
    types := [...]string{"Whitespace", "Ident", "String", "Comment",
        "Int", "Float", "Symbol"}
    if int(t) > len(types) - 1 {
        return ""
    }

    return types[t]
}

// Represents the position of the current token.
type Position struct {
    Filename string // Filename, if any.
    Offset int      // Byte offset (starting at 0).
    Line int        // Line number (starting at 1).
    Column int      // Column number (in characters, starting at 1).
}

// Returns a string representation of the current position.
func (p *Position) String() string {
    return fmt.Sprintf("%s:%d:%d (%d)", p.Filename, p.Line, p.Column,
        p.Offset)
}

// A Token.
type Token struct {
    Text string    // The text of the token.
    NumBytes int   // Number of bytes in the token.
    NumChars int   // Number of characters/runes in the token.
    FirstRune rune // First rune in the token.
    Type TokenType // The type of token.
}

func (t *Token) String() string {
    s := fmt.Sprintf("t=%s r=%c nc=%d nb=%d: %q", t.Type, t.FirstRune,
        t.NumChars, t.NumBytes, t.Text)
    return s
}

// A TokenScanner.
type TokenScanner struct {
    filename string
    reader *bufio.Reader
    pos *Position
    last_err error
    last_byte_len int
    last_line_addition int
    last_col int
    eol rune

    // Indicator to skip whitespace tokens.
    SkipWhitespace bool

    // Indicator to skip comment tokens.
    SkipComments bool

    // The most recent Token generated by a call to Scan().
    LastToken *Token

    // Predicate controlling the characters accepted as the i'th rune in an
    // identifier (starting at zero). The set of valid characters must not
    // intersect with the set of white space characters. The default is the
    // IsIdentRune function defined in this module.
    IsIdentRune func(ch rune, i int) bool

    // Predicate controlling the characters accepted as white space. The
    // default value is `unicode.IsSpace()`, which decides based on Unicode's
    // White Space property.
    IsSpaceRune func(ch rune) bool

    // Predicate controlling the characters accepted as quoting runes. Returns
    // true/false, as well as the corresponding closing quote rune. The
    // default is the IsQuoteRune define in this module.
    IsQuoteRune func(ch rune) (bool, rune)

    // Predicate controlling the characters accepted as escape runes, e.g.,
    // for escaping a quote character inside quotes.
    IsEscapeRune func(ch rune) (bool)

    // Predicate controlling the characters accepted as symbols.
    IsSymbolRune func(ch rune) bool

    // Predicate controlling the characters accepted as numeric digits.
    IsDigitRune func(ch rune) bool
}

// Returns position information for the current state. The same Position
// object is used throughout parsing.
func (ts *TokenScanner) Position() *Position {
    return ts.pos
}

// Returns a new TokenScanner initialized with the provided reader.
func NewScanner(r io.Reader) *TokenScanner {
    ts := new(TokenScanner)
    ts.Init(r)
    return ts
}

// Returns a TokenScanner initialized with the contents of the provided
// string.
func NewScannerString(s string) *TokenScanner {
    return NewScanner(strings.NewReader(s))
}

// Returns a TokenScanner initialized with the contents of the provided
// byte slice.
func NewScannerBytes(b []byte) *TokenScanner {
    return NewScanner(bytes.NewReader(b))
}

// Initializes a TokenScanner with the provided reader. This is only needed if
// a TokenScanner is created outside of one of the New* functions.
func (ts *TokenScanner) Init(r io.Reader) {
    ts.reader = bufio.NewReader(r)
    ts.pos = &Position{
        Line: 1,
        Column: 1,
    }

    ts.IsIdentRune = IsIdentRune
    ts.IsSpaceRune = unicode.IsSpace
    ts.IsQuoteRune = IsQuoteRune
    ts.IsEscapeRune = IsEscapeRune
    ts.IsSymbolRune = IsSymbolRune
    ts.IsDigitRune = unicode.IsDigit

    ts.SkipWhitespace = true
    ts.SkipComments = true

    ts.last_byte_len = 0
    ts.last_line_addition = 0
    ts.last_col = 1

    ts.eol = '\n'
}

// Returns the last error encountered.
func (ts *TokenScanner) Err() error {
    return ts.last_err
}

// Returns the most recent token generated by a call to Scan().
func (ts *TokenScanner) Token() *Token {
    return ts.LastToken
}

// Returns the text from the most recent token generated by a call to Scan().
func (ts *TokenScanner) TokenText() string {
    if ts.LastToken == nil {
        return ""
    }

    return ts.LastToken.Text
}

// Sets the rune considered to be the end-of-line character.
func (ts *TokenScanner) SetEOL(eol rune) {
    ts.eol = eol
}

// Sets the file name returned in the Position object.
func (ts *TokenScanner) SetFilename(filename string) {
    ts.pos.Filename = filename
}

func (ts *TokenScanner) update_pos() {
    pos := ts.pos

    // Add the byte length of the last token.
    pos.Offset += ts.last_byte_len
    ts.last_byte_len = 0

    // Add any additional lines parsed in the last token.
    pos.Line += ts.last_line_addition
    ts.last_line_addition = 0

    // Set to the last column count. `last_col` gets reset to 1 anytime the
    // end-of-line character is found.
    pos.Column = ts.last_col
}

// Scans the next token, skipping whitespace and comments, unless configured
// differently. Returns true if another token was found. Returns false when
// parsing is completed. Check ts.Err() for parsing errors.
func (ts *TokenScanner) Scan() bool {
    var (
        done bool
        err error
        token *Token
    )

    defer func() { ts.last_err = err }()

    for !done {
        ts.update_pos()

        token, err = ts.get_whitespace()
        if token != nil {
            if ts.SkipWhitespace {
                continue
            }
            return true
        }
        if err != nil {
            return false
        }

        token, err = ts.get_comment()
        if token != nil {
            if ts.SkipComments {
                continue
            }
            return true
        }
        if err != nil {
            return false
        }

        token, err = ts.get_quoted()
        if token != nil {
            return true
        }
        if err != nil {
            return false
        }

        token, err = ts.get_ident()
        if token != nil {
            return true
        }
        if err != nil {
            return false
        }

        token, err = ts.get_number()
        if token != nil {
            return true
        }
        if err != nil {
            return false
        }

        token, err = ts.get_symbol()
        if token != nil {
            return true
        }
        if err != nil {
            return false
        }

        done = true
    }

    return false
}

func (ts *TokenScanner) check_next_rune_char(ch rune) bool {
    next_ch, err := ts.peek_rune()
    if err != nil {
        return false
    }

    if next_ch == ch {
        return true
    }

    return false
}

func (ts *TokenScanner) check_next_rune_char_n(ch rune, n int) bool {
    runes, err := ts.peek_multirune(n)
    if err != nil {
        return false
    }

    if len(runes) < n {
        return false
    }

    next_ch := runes[n - 1]

    if next_ch == ch {
        return true
    }

    return false
}

func (ts *TokenScanner) check_next_rune_class(check func(rune) bool) bool {
    next_ch, err := ts.peek_rune()
    if err != nil {
        return false
    }

    if check(next_ch) {
        return true
    }

    return false
}

func (ts *TokenScanner) check_next_rune_class_n(
    check func(rune) bool,
    n int,
) bool {
    runes, err := ts.peek_multirune(n)
    if err != nil {
        return false
    }

    if len(runes) < n {
        return false
    }

    ch := runes[n - 1]

    if check(ch) {
        return true
    }

    return false
}

func (ts *TokenScanner) peek_rune() (rune, error) {
    runes, err := ts.peek_multirune(1)
    if err != nil {
        return 0, err
    }
    return runes[0], nil
}

func (ts *TokenScanner) peek_multirune(num_runes int) ([]rune, error) {
    buf, err := ts.reader.Peek(4 * num_runes)
    if err != nil {
        if ! (err == io.EOF && len(buf) > 0) {
            return nil, err
        }
    }

    runes := make([]rune, 0, num_runes)
    offset := 0

    for i := 0; i < num_runes; i ++ {
        ch, size := utf8.DecodeRune(buf[offset:])
        if size == 0 {
            return nil, io.EOF
        }

        offset += size

        if ch == utf8.RuneError {
            return runes, fmt.Errorf("invalid utf-8 sequence")
        }

        runes = append(runes, ch)
    }

    return runes, nil
}

func (ts *TokenScanner) get_ident() (*Token, error) {
    var (
        runes []rune
        total_size int
    )

    for i := 0; true; i++ {
        ch, size, err := ts.get_one_rune()
        if err != nil {
            if err == io.EOF && len(runes) > 0 {
                break
            }
            return nil, err
        }

        if ts.IsIdentRune(ch, i) {
            total_size += size
            if ch == ts.eol {
                ts.last_line_addition++
                ts.last_col = 1
            } else {
                ts.last_col++
            }

            runes = append(runes, ch)
            continue
        }

        if err = ts.unread_rune(); err != nil {
            return nil, nil
        }

        break
    }

    if len(runes) == 0 {
        return nil, nil
    }

    b := new(strings.Builder)
    for _, r := range runes {
        b.WriteRune(r)
    }

    text := b.String()
    token := &Token{
        Text: text,
        NumBytes: total_size,
        NumChars: len(runes),
        FirstRune: runes[0],
        Type: TokenTypeIdent,
    }

    ts.last_byte_len = total_size
    ts.LastToken = token

    return token, nil
}

func (ts *TokenScanner) read_until(end_ch rune) ([]rune, error) {
    var runes []rune

    for {
        ch, size, err := ts.get_one_rune()
        if err != nil {
            // Special case for EOF when we're looking for an end-of-line
            if err == io.EOF && end_ch == ts.eol {
                break
            }
            return nil, err
        }

        ts.last_byte_len += size

        if ch == ts.eol {
            ts.last_line_addition++
            ts.last_col = 1
        } else {
            ts.last_col++
        }

        runes = append(runes, ch)

        if ch == end_ch {
            break
        }
    }

    if len(runes) == 0 {
        return nil, nil
    }

    return runes, nil
}

func (ts *TokenScanner) get_comment() (*Token, error) {
    ch, _, err := ts.get_one_rune()
    if err != nil {
        return nil, err
    }

    if ch == '/' {
        if err = ts.unread_rune(); err != nil {
            return nil, err
        }

        var all_runes []rune

        if ts.check_next_rune_char_n('/', 2) {
            // This is a line comment.
            chars, _, err := ts.get_n_runes(2)
            if err != nil {
                return nil, err
            }

            all_runes = append(all_runes, chars...)

            chars, err = ts.read_until(ts.eol)
            if err != nil && err != io.EOF {
                return nil, err
            }

            all_runes = append(all_runes, chars...)

        } else if ts.check_next_rune_char_n('*', 2) {
            // This is a multi-line comment.
            chars, _, err := ts.get_n_runes(2)
            if err != nil {
                return nil, err
            }

            all_runes = append(all_runes, chars...)

            done := false
            for !done {
                done = true
                runes, err := ts.read_until('*')
                if err != nil {
                    return nil, err
                }
                all_runes = append(all_runes, runes...)

                ch, _, err = ts.get_one_rune()
                if err != nil {
                    return nil, err
                }
                all_runes = append(all_runes, ch)
                if ch != '/' {
                    // keep going
                    done = false
                }
            }
        }

        if len(all_runes) > 0 {
            token := &Token{
                Text: runes_to_string(all_runes),
                NumBytes: ts.last_byte_len,
                NumChars: len(all_runes),
                FirstRune: '/',
                Type: TokenTypeComment,
            }

            ts.LastToken = token

            return token, nil
        }

        return nil, nil
    }

    if err = ts.unread_rune(); err != nil {
        return nil, err
    }

    return nil, nil
}

func (ts *TokenScanner) get_quoted() (*Token, error) {
    ch, size, err := ts.get_one_rune()
    if err != nil {
        return nil, err
    }

    ok, closing_char := ts.IsQuoteRune(ch)
    if !ok {
        if err = ts.unread_rune(); err != nil {
            return nil, err
        }
        return nil, nil
    }

    ts.last_byte_len += size

    all_runes := []rune{}

    done := true
    loop_num := 0
    for {
        done = true
        loop_num++
        runes, err := ts.read_until(closing_char)
        if err != nil {
            return nil, fmt.Errorf("Unterminated string at %s. Couldn't " +
                "find end quote (%c).", ts.Position(), closing_char)
        }

        if len(runes) > 1 {
            i := len(runes) - 1 // last element
            if ts.IsEscapeRune(runes[i - 1]) {
                // Overwrite the escape character with the last character and
                // truncate.
                runes = append(runes[:i - 1], runes[i])

                // Make sure we loop again to get the rest of the quoted
                // string.
                done = false
            }
        }

        all_runes = append(all_runes, runes...)
        if done {
            break
        }
    }

    text := runes_to_string([]rune{ch}, all_runes)

    token := &Token{
        Text: text,
        NumBytes: ts.last_byte_len,
        NumChars: len(all_runes) + 1,
        FirstRune: ch,
        Type: TokenTypeString,
    }

    ts.LastToken = token

    return token, nil
}

type predicate_func func(rune) bool

func (ts *TokenScanner) get_general(
    token_type TokenType,
    rune_check predicate_func,
    exceptions ...predicate_func,
) (*Token, error) {
    var (
        runes []rune
        total_size int
    )

    for {
        ch, size, err := ts.get_one_rune()
        if err != nil {
            if err == io.EOF && len(runes) > 0 {
                break
            }
            return nil, err
        }

        is_exception := false
        for _, e := range exceptions {
            if e(ch) {
                is_exception = true
                break
            }
        }

        if !is_exception {
            if rune_check(ch) {
                total_size += size
                if ch == ts.eol {
                    ts.last_line_addition++
                    ts.last_col = 1
                } else {
                    ts.last_col++
                }

                runes = append(runes, ch)
                continue
            }
        }

        if err = ts.unread_rune(); err != nil {
            return nil, err
        }

        break
    }

    if len(runes) == 0 {
        return nil, nil
    }

    text := runes_to_string(runes)

    token := &Token{
        Text: text,
        NumBytes: total_size,
        NumChars: len(runes),
        FirstRune: runes[0],
        Type: token_type,
    }

    ts.last_byte_len = total_size
    ts.LastToken = token

    return token, nil
}

func runes_to_string(args ...[]rune) string {
    b := new(strings.Builder)

    for _, runes := range args {
        for _, r := range runes {
            b.WriteRune(r)
        }
    }

    return b.String()
}

func (ts *TokenScanner) get_number() (*Token, error) {
    var (
        runes []rune
        total_size int
    )

    found_digits := false
    found_decimal := false
    is_float := false

    for {
        ch, size, err := ts.get_one_rune()
        if err != nil {
            if err == io.EOF && len(runes) > 0 {
                break
            }
            return nil, err
        }

        if ch == '.' {
            if found_digits && !found_decimal {
                // We can't unread a rune after peeking ahead. So we unread
                // the rune here, then peek two runes ahead to see if the
                // period is followed by a digit. If so, read in the period
                // again to set us up for the next loop iteration. If not,
                // break out of the loop, as we've reached the end of the
                // number.
                if err = ts.unread_rune(); err != nil {
                    return nil, err
                }

                // Check if there is a digit after the decimal to determine if
                // we're reading floating point number or this is just a
                // period at the end of an integer.
                if ts.check_next_rune_class_n(ts.IsDigitRune, 2) {
                    found_decimal = true
                    is_float = true
                    total_size += size
                    ts.last_col++
                    runes = append(runes, ch)

                    // Read the period back in and continue on.
                    ch, size, err = ts.get_one_rune()
                    if err != nil {
                        if err == io.EOF && len(runes) > 0 {
                            break
                        }
                        return nil, err
                    }
                    continue
                } else {
                    break
                }
            }
        }

        if ch == '-' {
            if !found_digits {
                if err = ts.unread_rune(); err != nil {
                    return nil, err
                }

                // Check if there is a digit after the minus sign to determine
                // if we're reading umber or this is just a a minus sign.
                if ts.check_next_rune_class_n(ts.IsDigitRune, 2) {
                    total_size += size
                    ts.last_col++
                    runes = append(runes, ch)

                    // Read back in the minus sign and continue
                    ch, size, err = ts.get_one_rune()
                    if err != nil {
                        if err == io.EOF && len(runes) > 0 {
                            break
                        }
                        return nil, err
                    }
                    continue
                } else {
                    break
                }
            }
        }

        if ts.IsDigitRune(ch) {
            found_digits = true
            total_size += size
            if ch == ts.eol {
                ts.last_line_addition++
                ts.last_col = 1
            } else {
                ts.last_col++
            }

            runes = append(runes, ch)
            continue
        }

        if err = ts.unread_rune(); err != nil {
            return nil, err
        }

        break
    }

    if len(runes) == 0 {
        return nil, nil
    }

    text := runes_to_string(runes)

    token_type := TokenTypeInt
    if is_float {
        token_type = TokenTypeFloat
    }

    token := &Token{
        Text: text,
        NumBytes: total_size,
        NumChars: len(runes),
        FirstRune: runes[0],
        Type: token_type,
    }

    ts.last_byte_len = total_size
    ts.LastToken = token

    return token, nil
}

func (ts *TokenScanner) get_symbol() (*Token, error) {
    quote_func := func(ch rune) bool {
        if ok, _ := ts.IsQuoteRune(ch); ok {
            return true
        }
        return false
    }
    return ts.get_general(TokenTypeSymbol, ts.IsSymbolRune, quote_func)
}

func (ts *TokenScanner) get_whitespace() (*Token, error) {
    return ts.get_general(TokenTypeWhitespace, ts.IsSpaceRune)
}

func (ts *TokenScanner) unread_rune() error {
    return ts.reader.UnreadRune()
}

func (ts *TokenScanner) get_n_runes(
    n int,
) (
    chars []rune,
    total_size int,
    err error,
) {
    var (
        ch rune
        size int
    )

    for i := 0; i < n; i++ {
        ch, size, err = ts.reader.ReadRune()
        if err != nil {
            ts.last_err = err
            return
        }
        chars = append(chars, ch)
        total_size += size

        ts.last_byte_len += size
        ts.last_col++
        if ch == ts.eol {
            ts.last_line_addition++
            ts.last_col = 1
        }
    }

    return
}

func (ts *TokenScanner) get_one_rune() (ch rune, size int, err error) {
    ch, size, err = ts.reader.ReadRune()
    if err != nil {
        ts.last_err = err
        return
    }

    return
}
