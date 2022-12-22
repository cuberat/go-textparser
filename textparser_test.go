package textparser_test

import (
	"fmt"
	textparser "github.com/cuberat/go-textparser"
	"io"
	"reflect"
	"strings"
	"testing"
)

type TestData struct {
	Name              string
	Input             string
	Expected          []string
	ExpectedTokens    []*textparser.Token
	ExpectedPositions []*textparser.Position
}

func TestSkipWhitespace(t *testing.T) {
	txt := "foo bar"
	p := new(textparser.TokenScanner)
	p.Init(strings.NewReader(txt))
	p.SkipWhitespace = true

	expected := []string{"foo", "bar"}
	token_list := make([]string, 0, len(expected))

	for p.Scan() {
		token_list = append(token_list, p.TokenText())
	}

	if !reflect.DeepEqual(expected, token_list) {
		t.Errorf("got %+v, expected %+v", token_list, expected)
	}
}

func fooTestStrings(t *testing.T) {
	txt := `if foo / bar "more stuff" and 'even more'`
	p := new(textparser.TokenScanner)
	p.Init(strings.NewReader(txt))
	p.SkipWhitespace = true

	expected := []string{"if", "foo", "/", "bar", `"more stuff"`, "and",
		"'even more'"}
	token_list := make([]string, 0, len(expected))

	for p.Scan() {
		token_list = append(token_list, p.TokenText())
	}

	if !reflect.DeepEqual(expected, token_list) {
		t.Errorf("got %+v, expected %+v", token_list, expected)
	}
}

func TestTagLike(t *testing.T) {
	txt := `name, del=',', usage='Use it this way. It\'s got stuff.'`
	p := new(textparser.TokenScanner)
	p.Init(strings.NewReader(txt))
	p.SkipWhitespace = true

	expected := []string{"name", ",", "del", "=", "','", ",", "usage", "=",
		`'Use it this way. It's got stuff.'`}
	token_list := make([]string, 0, len(expected))

	for p.Scan() {
		token_list = append(token_list, p.TokenText())
	}

	if err := p.Err(); err != nil {
		if err != io.EOF {
			t.Errorf("error from scanner: %s", err)
			return
		}
	}

	if !reflect.DeepEqual(expected, token_list) {
		t.Errorf("got %#v, expected %#v", token_list, expected)
	}
}

// Test that an error is returned if a quoted string is unterminated.
func TestQuotedUnterminated(t *testing.T) {
	var err error
	tests := []*TestData{
		&TestData{
			Name:  `quoted (")`,
			Input: `foo = "foo bar`,
		},
		&TestData{
			Name:  `quoted (“”)`,
			Input: `foo = “foo bar`,
		},
	}

	for _, test_data := range tests {
		t.Run(test_data.Name, func(st *testing.T) {
			p := new(textparser.TokenScanner)
			p.Init(strings.NewReader(test_data.Input))
			p.IsQuoteRune = textparser.IsQuoteRuneFancy

			token_list := make([]string, 0, len(test_data.Expected))

			for p.Scan() {
				token_list = append(token_list, p.TokenText())
			}

			if err = p.Err(); err == nil || err == io.EOF {
				st.Errorf("expected error for unterminated string %q at: %s", test_data.Input, p.Position())
				return
			}
		})
	}
}

func TestNumbers(t *testing.T) {
	tests := []*TestData{
		&TestData{
			Name:     `integers`,
			Input:    `foo = 42`,
			Expected: []string{"foo", "=", `42`},
		},
		&TestData{
			Name:     `negative integers`,
			Input:    `foo = -42`,
			Expected: []string{"foo", "=", `-42`},
		},

		&TestData{
			Name:     `floats`,
			Input:    `foo = 42.5`,
			Expected: []string{"foo", "=", `42.5`},
		},
		&TestData{
			Name:     `negative floats`,
			Input:    `foo = -42.5`,
			Expected: []string{"foo", "=", `-42.5`},
		},

		&TestData{
			Name:     `integers ending in period`,
			Input:    `foo = 42.`,
			Expected: []string{"foo", "=", `42`, `.`},
		},
		&TestData{
			Name:     `negative integers ending in period`,
			Input:    `foo = -42.`,
			Expected: []string{"foo", "=", `-42`, `.`},
		},
	}

	for _, test_data := range tests {
		t.Run(test_data.Name, func(st *testing.T) {
			p := new(textparser.TokenScanner)
			p.Init(strings.NewReader(test_data.Input))
			token_list := make([]string, 0, len(test_data.Expected))

			for p.Scan() {
				token_list = append(token_list, p.TokenText())
			}

			if err := p.Err(); err != nil {
				if err != io.EOF {
					st.Errorf("error from scanner: %s", err)
					return
				}
			}

			if !reflect.DeepEqual(test_data.Expected, token_list) {
				st.Errorf("got %#v, expected %#v",
					token_list, test_data.Expected)
			}
		})
	}
}

func TestComments(t *testing.T) {
	tests := []*TestData{
		&TestData{
			Name:     `line comment (//)`,
			Input:    `foo = // h4x0r and stuff`,
			Expected: []string{"foo", "=", `// h4x0r and stuff`},
		},
		&TestData{
			Name:     `multi-line comment (/* */)`,
			Input:    `foo = /* h4x0r and stuff */`,
			Expected: []string{"foo", "=", `/* h4x0r and stuff */`},
		},
		&TestData{
			Name:     `multi-line comment with fake (/* * */)`,
			Input:    `foo = /* h4x0r and * stuff */`,
			Expected: []string{"foo", "=", `/* h4x0r and * stuff */`},
		},
	}

	for _, test_data := range tests {
		t.Run(test_data.Name, func(st *testing.T) {
			p := new(textparser.TokenScanner)
			p.Init(strings.NewReader(test_data.Input))
			p.SkipWhitespace = true
			p.SkipComments = false

			token_list := make([]string, 0, len(test_data.Expected))

			for p.Scan() {
				token_list = append(token_list, p.TokenText())
			}

			if err := p.Err(); err != nil {
				if err != io.EOF {
					st.Errorf("error from scanner: %s", err)
					return
				}
			}

			if !reflect.DeepEqual(test_data.Expected, token_list) {
				st.Errorf("got %#v, expected %#v",
					token_list, test_data.Expected)
			}
		})
	}
}

func TestSkipComments(t *testing.T) {
	tests := []*TestData{
		&TestData{
			Name:     `line comment (//)`,
			Input:    `foo = // h4x0r and stuff`,
			Expected: []string{"foo", "="},
		},
		&TestData{
			Name:     `multi-line comment (/* */)`,
			Input:    `foo = /* h4x0r and stuff */`,
			Expected: []string{"foo", "="},
		},
		&TestData{
			Name:     `multi-line comment with fake (/* * */)`,
			Input:    `foo = /* h4x0r and * stuff */`,
			Expected: []string{"foo", "="},
		},
	}

	for _, test_data := range tests {
		t.Run(test_data.Name, func(st *testing.T) {
			p := new(textparser.TokenScanner)
			p.Init(strings.NewReader(test_data.Input))
			p.SkipWhitespace = true
			p.SkipComments = true

			token_list := make([]string, 0, len(test_data.Expected))

			for p.Scan() {
				token_list = append(token_list, p.TokenText())
			}

			if err := p.Err(); err != nil {
				if err != io.EOF {
					st.Errorf("error from scanner: %s", err)
					return
				}
			}

			if !reflect.DeepEqual(test_data.Expected, token_list) {
				st.Errorf("got %#v, expected %#v",
					token_list, test_data.Expected)
			}
		})
	}
}

func TestQuoted(t *testing.T) {
	tests := []*TestData{
		&TestData{
			Name:     `quoted (")`,
			Input:    `foo = "foo bar"`,
			Expected: []string{"foo", "=", `"foo bar"`},
		},

		&TestData{
			Name:     `quoted (") with escaped quotes`,
			Input:    `foo = "foo \"some stuff\" bar"`,
			Expected: []string{"foo", "=", `"foo "some stuff" bar"`},
		},

		&TestData{
			Name:     `quoted (')`,
			Input:    `foo = 'foo bar'`,
			Expected: []string{"foo", "=", `'foo bar'`},
		},

		&TestData{
			Name:     `quoted (') with escaped quotes`,
			Input:    `foo = 'foo \'some stuff\' bar'`,
			Expected: []string{"foo", "=", `'foo 'some stuff' bar'`},
		},

		&TestData{
			Name:     "quoted (`)",
			Input:    "foo = `foo bar`",
			Expected: []string{"foo", "=", "`foo bar`"},
		},

		&TestData{
			Name:     "quoted (`) with escaped quotes",
			Input:    "foo = `foo \\`some stuff\\` bar`",
			Expected: []string{"foo", "=", "`foo `some stuff` bar`"},
		},
	}

	for _, test_data := range tests {
		t.Run(test_data.Name, func(st *testing.T) {
			p := new(textparser.TokenScanner)
			p.Init(strings.NewReader(test_data.Input))
			token_list := make([]string, 0, len(test_data.Expected))

			for p.Scan() {
				token_list = append(token_list, p.TokenText())
			}

			if err := p.Err(); err != nil {
				if err != io.EOF {
					st.Errorf("error from scanner: %s", err)
					return
				}
			}

			if !reflect.DeepEqual(test_data.Expected, token_list) {
				st.Errorf("got %#v, expected %#v",
					token_list, test_data.Expected)
			}
		})
	}
}

func TestQuotedFancy(t *testing.T) {
	tests := []*TestData{
		&TestData{
			Name:     `quoted (")`,
			Input:    `foo = "foo bar"`,
			Expected: []string{"foo", "=", `"foo bar"`},
		},

		&TestData{
			Name:     `quoted (") with escaped quotes`,
			Input:    `foo = "foo \"some stuff\" bar"`,
			Expected: []string{"foo", "=", `"foo "some stuff" bar"`},
		},

		&TestData{
			Name:     `quoted (')`,
			Input:    `foo = 'foo bar'`,
			Expected: []string{"foo", "=", `'foo bar'`},
		},

		&TestData{
			Name:     `quoted (') with escaped quotes`,
			Input:    `foo = 'foo \'some stuff\' bar'`,
			Expected: []string{"foo", "=", `'foo 'some stuff' bar'`},
		},

		&TestData{
			Name:     "quoted (`)",
			Input:    "foo = `foo bar`",
			Expected: []string{"foo", "=", "`foo bar`"},
		},

		&TestData{
			Name:     "quoted (`) with escaped quotes",
			Input:    "foo = `foo \\`some stuff\\` bar`",
			Expected: []string{"foo", "=", "`foo `some stuff` bar`"},
		},

		// Fancy quotes from here down.
		&TestData{
			Name:     `quoted (“”)`,
			Input:    `foo = “foo bar”`,
			Expected: []string{"foo", "=", `“foo bar”`},
		},

		&TestData{
			Name:     `quoted (“”) with escaped quotes`,
			Input:    `foo = “foo “some stuff\” bar”`,
			Expected: []string{"foo", "=", `“foo “some stuff” bar”`},
		},

		&TestData{
			Name:     `quoted (‘’)`,
			Input:    `foo = ‘foo bar’`,
			Expected: []string{"foo", "=", `‘foo bar’`},
		},

		&TestData{
			Name:     `quoted (‘’) with escaped quotes`,
			Input:    `foo = ‘foo ‘some stuff\’ bar’`,
			Expected: []string{"foo", "=", `‘foo ‘some stuff’ bar’`},
		},

		&TestData{
			Name:     `quoted (‹›)`,
			Input:    `foo = ‹foo bar›`,
			Expected: []string{"foo", "=", `‹foo bar›`},
		},

		&TestData{
			Name:     `quoted (‹›) with escaped quotes`,
			Input:    `foo = ‹foo ‹some stuff\› bar›`,
			Expected: []string{"foo", "=", `‹foo ‹some stuff› bar›`},
		},

		&TestData{
			Name:     `quoted («»)`,
			Input:    `foo = «foo bar»`,
			Expected: []string{"foo", "=", `«foo bar»`},
		},

		&TestData{
			Name:     `quoted («») with escaped quotes`,
			Input:    `foo = «foo «some stuff\» bar»`,
			Expected: []string{"foo", "=", `«foo «some stuff» bar»`},
		},
	}

	for _, test_data := range tests {
		t.Run(test_data.Name, func(st *testing.T) {
			p := new(textparser.TokenScanner)
			p.Init(strings.NewReader(test_data.Input))
			p.IsQuoteRune = textparser.IsQuoteRuneFancy

			token_list := make([]string, 0, len(test_data.Expected))

			for p.Scan() {
				token_list = append(token_list, p.TokenText())
			}

			if err := p.Err(); err != nil {
				if err != io.EOF {
					st.Errorf("error from scanner: %s", err)
					return
				}
			}

			if !reflect.DeepEqual(test_data.Expected, token_list) {
				st.Errorf("got %#v, expected %#v",
					token_list, test_data.Expected)
			}
		})
	}
}

func TestPosition(t *testing.T) {
	txt := "foo = 'bar'\nbar = 'foo'"
	expected := []string{"foo", "=", "'bar'", "bar", "=", "'foo'"}
	expected_pos := []*textparser.Position{
		&textparser.Position{
			Filename: "test_file",
			Offset:   0,
			Line:     1,
			Column:   1,
		},

		&textparser.Position{
			Filename: "test_file",
			Offset:   4,
			Line:     1,
			Column:   5,
		},

		&textparser.Position{
			Filename: "test_file",
			Offset:   6,
			Line:     1,
			Column:   7,
		},

		&textparser.Position{
			Filename: "test_file",
			Offset:   12,
			Line:     2,
			Column:   1,
		},

		&textparser.Position{
			Filename: "test_file",
			Offset:   16,
			Line:     2,
			Column:   5,
		},

		&textparser.Position{
			Filename: "test_file",
			Offset:   18,
			Line:     2,
			Column:   7,
		},
	}

	p := new(textparser.TokenScanner)
	p.Init(strings.NewReader(txt))
	p.SetFilename("test_file")

	token_list := make([]string, 0, len(expected))

	idx := -1
	for p.Scan() {
		idx++
		token_list = append(token_list, p.TokenText())

		pos := p.Position()
		if !reflect.DeepEqual(pos, expected_pos[idx]) {
			t.Errorf("token %q: got %s, expected %s",
				p.TokenText(), pos, expected_pos[idx])
		}
	}
}

func TestPositionEmbeddedEOL(t *testing.T) {
	txt := "foo = 'bar\nstuff' and"
	expected := []string{"foo", "=", "'bar\nstuff' and"}
	expected_pos := []*textparser.Position{
		&textparser.Position{
			Filename: "test_file",
			Offset:   0,
			Line:     1,
			Column:   1,
		},

		&textparser.Position{
			Filename: "test_file",
			Offset:   4,
			Line:     1,
			Column:   5,
		},

		&textparser.Position{
			Filename: "test_file",
			Offset:   6,
			Line:     1,
			Column:   7,
		},

		&textparser.Position{
			Filename: "test_file",
			Offset:   18,
			Line:     2,
			Column:   8,
		},
	}

	p := new(textparser.TokenScanner)
	p.Init(strings.NewReader(txt))
	p.SetFilename("test_file")

	token_list := make([]string, 0, len(expected))

	idx := -1
	for p.Scan() {
		idx++
		token_list = append(token_list, p.TokenText())

		t.Run(fmt.Sprintf("%s", expected_pos[idx]), func(st *testing.T) {
			pos := p.Position()
			if !reflect.DeepEqual(pos, expected_pos[idx]) {
				st.Errorf("token %q: got %s, expected %s",
					p.TokenText(), pos, expected_pos[idx])
			}
		})
	}
}

func TestTokens(t *testing.T) {
	tests := []*TestData{
		&TestData{
			Name:     `line comment (//)`,
			Input:    `foo = // h4x0r and stuff`,
			Expected: []string{"foo", "=", `// h4x0r and stuff`},
			ExpectedTokens: []*textparser.Token{
				&textparser.Token{
					Text:      "foo",
					NumBytes:  3,
					NumChars:  3,
					FirstRune: 'f',
					Type:      textparser.TokenTypeIdent,
				},
				&textparser.Token{
					Text:      " ",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: ' ',
					Type:      textparser.TokenTypeWhitespace,
				},
				&textparser.Token{
					Text:      "=",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: '=',
					Type:      textparser.TokenTypeSymbol,
				},
				&textparser.Token{
					Text:      " ",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: ' ',
					Type:      textparser.TokenTypeWhitespace,
				},
				&textparser.Token{
					Text:      `// h4x0r and stuff`,
					NumBytes:  18,
					NumChars:  18,
					FirstRune: '/',
					Type:      textparser.TokenTypeComment,
				},
			},
		},

		&TestData{
			Name:  `numbers`,
			Input: `5 42.5`,
			ExpectedTokens: []*textparser.Token{
				&textparser.Token{
					Text:      "5",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: '5',
					Type:      textparser.TokenTypeInt,
				},
				&textparser.Token{
					Text:      " ",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: ' ',
					Type:      textparser.TokenTypeWhitespace,
				},
				&textparser.Token{
					Text:      "42.5",
					NumBytes:  4,
					NumChars:  4,
					FirstRune: '4',
					Type:      textparser.TokenTypeFloat,
				},
			},
		},
	}

	for _, test_data := range tests {
		t.Run(test_data.Name, func(st *testing.T) {
			p := new(textparser.TokenScanner)
			p.Init(strings.NewReader(test_data.Input))
			p.SkipWhitespace = false
			p.SkipComments = false

			token_list := make([]*textparser.Token, 0,
				len(test_data.ExpectedTokens))

			for p.Scan() {
				token_list = append(token_list, p.Token())
			}

			if err := p.Err(); err != nil {
				if err != io.EOF {
					st.Errorf("error from scanner: %s", err)
					return
				}
			}

			if !reflect.DeepEqual(test_data.ExpectedTokens, token_list) {
				st.Errorf("got %+v, expected %+v",
					token_list, test_data.ExpectedTokens)
			}
		})
	}
}

func TestSeparateSymbols(t *testing.T) {
	tests := []*TestData{
		&TestData{
			Name:     `symbols +=`,
			Input:    `foo += 5`,
			Expected: []string{"foo", "+", "=", "5"},
			ExpectedTokens: []*textparser.Token{
				&textparser.Token{
					Text:      "foo",
					NumBytes:  3,
					NumChars:  3,
					FirstRune: 'f',
					Type:      textparser.TokenTypeIdent,
				},
				&textparser.Token{
					Text:      "+",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: '+',
					Type:      textparser.TokenTypeSymbol,
				},
				&textparser.Token{
					Text:      "=",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: '=',
					Type:      textparser.TokenTypeSymbol,
				},
				&textparser.Token{
					Text:      "5",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: '5',
					Type:      textparser.TokenTypeInt,
				},
			},
		},
	}

	for _, test_data := range tests {
		t.Run(test_data.Name, func(st *testing.T) {
			p := new(textparser.TokenScanner)
			p.Init(strings.NewReader(test_data.Input))
			p.SkipWhitespace = true
			p.SkipComments = true

			token_list := make([]*textparser.Token, 0,
				len(test_data.ExpectedTokens))

			for p.Scan() {
				token_list = append(token_list, p.Token())
			}

			if err := p.Err(); err != nil {
				if err != io.EOF {
					st.Errorf("error from scanner: %s", err)
					return
				}
			}

			if !reflect.DeepEqual(test_data.ExpectedTokens, token_list) {
				st.Errorf("got %+v, expected %+v",
					token_list, test_data.ExpectedTokens)
			}
		})
	}
}

func TestSomeSeparateSymbols(t *testing.T) {
	tests := []*TestData{
		&TestData{
			Name:     `one symbol += rest separate`,
			Input:    `foo += 5 })`,
			Expected: []string{"foo", "+=", "5", "}", ")"},
			ExpectedTokens: []*textparser.Token{
				&textparser.Token{
					Text:      "foo",
					NumBytes:  3,
					NumChars:  3,
					FirstRune: 'f',
					Type:      textparser.TokenTypeIdent,
				},
				&textparser.Token{
					Text:      "+=",
					NumBytes:  2,
					NumChars:  2,
					FirstRune: '+',
					Type:      textparser.TokenTypeSymbol,
				},
				&textparser.Token{
					Text:      "5",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: '5',
					Type:      textparser.TokenTypeInt,
				},
				&textparser.Token{
					Text:      "}",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: '}',
					Type:      textparser.TokenTypeSymbol,
				},
				&textparser.Token{
					Text:      ")",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: ')',
					Type:      textparser.TokenTypeSymbol,
				},
			},
		},
	}

	for _, test_data := range tests {
		t.Run(test_data.Name, func(st *testing.T) {
			p := new(textparser.TokenScanner)
			p.Init(strings.NewReader(test_data.Input))
			p.SkipWhitespace = true
			p.SkipComments = true

			p.IsSymbolRune = func(ch rune, i int, runes []rune) bool {
				if ch == '=' && i == 1 && runes[0] == '+' {
					return true
				}

				return textparser.IsSymbolRune(ch, i, runes)
			}

			token_list := make([]*textparser.Token, 0,
				len(test_data.ExpectedTokens))

			for p.Scan() {
				token_list = append(token_list, p.Token())
			}

			if err := p.Err(); err != nil {
				if err != io.EOF {
					st.Errorf("error from scanner: %s", err)
					return
				}
			}

			if !reflect.DeepEqual(test_data.ExpectedTokens, token_list) {
				st.Errorf("got %+v, expected %+v",
					token_list, test_data.ExpectedTokens)
			}
		})
	}
}

func TestUnreadToken(t *testing.T) {
	filename := ""
	tests := []*TestData{
		&TestData{
			Name:     `UnreadToken`,
			Input:    `foo += 5`,
			Expected: []string{"foo", "+", "+", "=", "5"},
			ExpectedTokens: []*textparser.Token{
				&textparser.Token{
					Text:      "foo",
					NumBytes:  3,
					NumChars:  3,
					FirstRune: 'f',
					Type:      textparser.TokenTypeIdent,
				},
				&textparser.Token{
					Text:      "+",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: '+',
					Type:      textparser.TokenTypeSymbol,
				},
				&textparser.Token{
					Text:      "+",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: '+',
					Type:      textparser.TokenTypeSymbol,
				},
				&textparser.Token{
					Text:      "=",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: '=',
					Type:      textparser.TokenTypeSymbol,
				},
				&textparser.Token{
					Text:      "5",
					NumBytes:  1,
					NumChars:  1,
					FirstRune: '5',
					Type:      textparser.TokenTypeInt,
				},
			},
			ExpectedPositions: []*textparser.Position{
				&textparser.Position{
					Filename: filename, Offset: 0, Line: 1, Column: 1,
				},
				&textparser.Position{
					Filename: filename, Offset: 4, Line: 1, Column: 5,
				},
				&textparser.Position{
					Filename: filename, Offset: 4, Line: 1, Column: 5,
				},
				&textparser.Position{
					Filename: filename, Offset: 5, Line: 1, Column: 6,
				},
				&textparser.Position{
					Filename: filename, Offset: 7, Line: 1, Column: 8,
				},
			},
		},
	}

	for _, test_data := range tests {
		t.Run(test_data.Name, func(st *testing.T) {
			p := new(textparser.TokenScanner)
			p.Init(strings.NewReader(test_data.Input))
			p.SkipWhitespace = true
			p.SkipComments = true

			token_list := make([]*textparser.Token, 0,
				len(test_data.ExpectedTokens))

			if len(test_data.ExpectedTokens) !=
				len(test_data.ExpectedPositions) {
				t.Errorf("number of expected tokens != number of expected " +
					"positions")
				return
			}

			for i := 0; p.Scan(); i++ {
				if i > len(test_data.ExpectedTokens)-1 {
					t.Errorf("too many tokens: at token %q", p.TokenText())
					return
				}
				token_list = append(token_list, p.Token())
				if !reflect.DeepEqual(p.Position(),
					test_data.ExpectedPositions[i]) {
					t.Errorf("[%d] got pos %s, expected %s", i,
						p.Position(), test_data.ExpectedPositions[i])
					return
				}
				if i == 1 {
					p.UnreadToken()
				}
			}

			if err := p.Err(); err != nil {
				if err != io.EOF {
					st.Errorf("error from scanner: %s", err)
					return
				}
			}

			if !reflect.DeepEqual(test_data.ExpectedTokens, token_list) {
				st.Errorf("got %+v, expected %+v",
					token_list, test_data.ExpectedTokens)
			}
		})
	}
}

func ExampleSetVar() {
	src := `
    // This is a comment.
    if a > 5 {
        b = "this is a string";
        c = 7.2;
    }
    `
	s := textparser.NewScanner(strings.NewReader(src))
	s.SetFilename("nofile")

	for s.Scan() {
		if err := s.Err(); err != nil {
			panic(fmt.Sprintf("error at %s: %s", s.Position(), err))
		}
		token := s.Token()
		fmt.Printf("%-16s - %-6s -> %s\n", s.Position(), token.Type,
			token.Text)
	}

	// Output:
	// nofile:3:5 (31)  - Ident  -> if
	// nofile:3:8 (34)  - Ident  -> a
	// nofile:3:10 (36) - Symbol -> >
	// nofile:3:12 (38) - Int    -> 5
	// nofile:3:14 (40) - Symbol -> {
	// nofile:4:9 (50)  - Ident  -> b
	// nofile:4:11 (52) - Symbol -> =
	// nofile:4:13 (54) - String -> "this is a string"
	// nofile:4:30 (72) - Symbol -> ;
	// nofile:5:9 (82)  - Ident  -> c
	// nofile:5:11 (84) - Symbol -> =
	// nofile:5:13 (86) - Float  -> 7.2
	// nofile:5:16 (89) - Symbol -> ;
	// nofile:6:5 (95)  - Symbol -> }
}

func ExampleStructTag() {
	src := `Verbose,del=',',usage='Use it like this.'`
	s := textparser.NewScanner(strings.NewReader(src))
	s.SetFilename("")

	for s.Scan() {
		if err := s.Err(); err != nil {
			panic(fmt.Sprintf("error at %s: %s", s.Position(), err))
		}
		token := s.Token()
		fmt.Printf("%-16s - %-6s -> %s\n", s.Position(), token.Type,
			token.Text)
	}

	// Output:
	// :1:1 (0)         - Ident  -> Verbose
	// :1:8 (7)         - Symbol -> ,
	// :1:9 (8)         - Ident  -> del
	// :1:12 (11)       - Symbol -> =
	// :1:13 (12)       - String -> ','
	// :1:15 (15)       - Symbol -> ,
	// :1:16 (16)       - Ident  -> usage
	// :1:21 (21)       - Symbol -> =
	// :1:22 (22)       - String -> 'Use it like this.'
}

// Example with customized symbol tokenization.
func ExampleCustomSymbols() {
	input := "(foo += 5 +-4)"

	ts := textparser.NewScanner(strings.NewReader(input))
	ts.SkipWhitespace = true
	ts.SkipComments = true

	// `+=` is considered one symbol, but all other symbols are single
	// characters.
	ts.IsSymbolRune = func(ch rune, i int, runes []rune) bool {
		if ch == '=' && i == 1 && runes[0] == '+' {
			return true
		}

		return textparser.IsSymbolRune(ch, i, runes)
	}

	for ts.Scan() {
		if err := ts.Err(); err != nil {
			fmt.Printf("====> Error during scan: %s", err)
			break
		}

		fmt.Printf("%s\n", ts.TokenText())
	}

	// Output:
	// (
	// foo
	// +=
	// 5
	// +
	// -4
	// )
}
