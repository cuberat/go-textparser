package textparser_test

import (
    "fmt"
    "io"
    "reflect"
    "strings"
    textparser "github.com/cuberat/go-textparser"
    "testing"
)

type TestData struct {
    Name string
    Input string
    Expected []string
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
            Name: `quoted (")`,
            Input: `foo = "foo bar`,
        },
        &TestData{
            Name: `quoted (“”)`,
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
            Name: `integers`,
            Input: `foo = 42`,
            Expected: []string{"foo", "=", `42`},
        },
        &TestData{
            Name: `negative integers`,
            Input: `foo = -42`,
            Expected: []string{"foo", "=", `-42`},
        },

        &TestData{
            Name: `floats`,
            Input: `foo = 42.5`,
            Expected: []string{"foo", "=", `42.5`},
        },
        &TestData{
            Name: `negative floats`,
            Input: `foo = -42.5`,
            Expected: []string{"foo", "=", `-42.5`},
        },

        &TestData{
            Name: `integers ending in period`,
            Input: `foo = 42.`,
            Expected: []string{"foo", "=", `42`, `.`},
        },
        &TestData{
            Name: `negative integers ending in period`,
            Input: `foo = -42.`,
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
            Name: `line comment (//)`,
            Input: `foo = // h4x0r and stuff`,
            Expected: []string{"foo", "=", `// h4x0r and stuff`},
        },
        &TestData{
            Name: `multi-line comment (/* */)`,
            Input: `foo = /* h4x0r and stuff */`,
            Expected: []string{"foo", "=", `/* h4x0r and stuff */`},
        },
        &TestData{
            Name: `multi-line comment with fake (/* * */)`,
            Input: `foo = /* h4x0r and * stuff */`,
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
            Name: `line comment (//)`,
            Input: `foo = // h4x0r and stuff`,
            Expected: []string{"foo", "="},
        },
        &TestData{
            Name: `multi-line comment (/* */)`,
            Input: `foo = /* h4x0r and stuff */`,
            Expected: []string{"foo", "="},
        },
        &TestData{
            Name: `multi-line comment with fake (/* * */)`,
            Input: `foo = /* h4x0r and * stuff */`,
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
            Name: `quoted (")`,
            Input: `foo = "foo bar"`,
            Expected: []string{"foo", "=", `"foo bar"`},
        },

        &TestData{
            Name: `quoted (") with escaped quotes`,
            Input: `foo = "foo \"some stuff\" bar"`,
            Expected: []string{"foo", "=", `"foo "some stuff" bar"`},
        },

        &TestData{
            Name: `quoted (')`,
            Input: `foo = 'foo bar'`,
            Expected: []string{"foo", "=", `'foo bar'`},
        },

        &TestData{
            Name: `quoted (') with escaped quotes`,
            Input: `foo = 'foo \'some stuff\' bar'`,
            Expected: []string{"foo", "=", `'foo 'some stuff' bar'`},
        },

        &TestData{
            Name: "quoted (`)",
            Input: "foo = `foo bar`",
            Expected: []string{"foo", "=", "`foo bar`"},
        },

        &TestData{
            Name: "quoted (`) with escaped quotes",
            Input: "foo = `foo \\`some stuff\\` bar`",
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
            Name: `quoted (")`,
            Input: `foo = "foo bar"`,
            Expected: []string{"foo", "=", `"foo bar"`},
        },

        &TestData{
            Name: `quoted (") with escaped quotes`,
            Input: `foo = "foo \"some stuff\" bar"`,
            Expected: []string{"foo", "=", `"foo "some stuff" bar"`},
        },

        &TestData{
            Name: `quoted (')`,
            Input: `foo = 'foo bar'`,
            Expected: []string{"foo", "=", `'foo bar'`},
        },

        &TestData{
            Name: `quoted (') with escaped quotes`,
            Input: `foo = 'foo \'some stuff\' bar'`,
            Expected: []string{"foo", "=", `'foo 'some stuff' bar'`},
        },

        &TestData{
            Name: "quoted (`)",
            Input: "foo = `foo bar`",
            Expected: []string{"foo", "=", "`foo bar`"},
        },

        &TestData{
            Name: "quoted (`) with escaped quotes",
            Input: "foo = `foo \\`some stuff\\` bar`",
            Expected: []string{"foo", "=", "`foo `some stuff` bar`"},
        },

        // Fancy quotes from here down.
        &TestData{
            Name: `quoted (“”)`,
            Input: `foo = “foo bar”`,
            Expected: []string{"foo", "=", `“foo bar”`},
        },

        &TestData{
            Name: `quoted (“”) with escaped quotes`,
            Input: `foo = “foo “some stuff\” bar”`,
            Expected: []string{"foo", "=", `“foo “some stuff” bar”`},
        },

        &TestData{
            Name: `quoted (‘’)`,
            Input: `foo = ‘foo bar’`,
            Expected: []string{"foo", "=", `‘foo bar’`},
        },

        &TestData{
            Name: `quoted (‘’) with escaped quotes`,
            Input: `foo = ‘foo ‘some stuff\’ bar’`,
            Expected: []string{"foo", "=", `‘foo ‘some stuff’ bar’`},
        },

        &TestData{
            Name: `quoted (‹›)`,
            Input: `foo = ‹foo bar›`,
            Expected: []string{"foo", "=", `‹foo bar›`},
        },

        &TestData{
            Name: `quoted (‹›) with escaped quotes`,
            Input: `foo = ‹foo ‹some stuff\› bar›`,
            Expected: []string{"foo", "=", `‹foo ‹some stuff› bar›`},
        },

        &TestData{
            Name: `quoted («»)`,
            Input: `foo = «foo bar»`,
            Expected: []string{"foo", "=", `«foo bar»`},
        },

        &TestData{
            Name: `quoted («») with escaped quotes`,
            Input: `foo = «foo «some stuff\» bar»`,
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
            Offset: 0,
            Line: 1,
            Column: 1,
        },

        &textparser.Position{
            Filename: "test_file",
            Offset: 4,
            Line: 1,
            Column: 5,
        },

        &textparser.Position{
            Filename: "test_file",
            Offset: 6,
            Line: 1,
            Column: 7,
        },

        &textparser.Position{
            Filename: "test_file",
            Offset: 12,
            Line: 2,
            Column: 1,
        },

        &textparser.Position{
            Filename: "test_file",
            Offset: 16,
            Line: 2,
            Column: 5,
        },

        &textparser.Position{
            Filename: "test_file",
            Offset: 18,
            Line: 2,
            Column: 7,
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
            Offset: 0,
            Line: 1,
            Column: 1,
        },

        &textparser.Position{
            Filename: "test_file",
            Offset: 4,
            Line: 1,
            Column: 5,
        },

        &textparser.Position{
            Filename: "test_file",
            Offset: 6,
            Line: 1,
            Column: 7,
        },

        &textparser.Position{
            Filename: "test_file",
            Offset: 18,
            Line: 2,
            Column: 8,
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
