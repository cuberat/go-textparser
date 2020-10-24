

# textparser
`import "github.com/cuberat/go-textparser"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>
Package textparser provides a scanner and tokenizer for UTF-8-encoded text.
It takes an io.Reader providing the source, which then can be tokenized
through repeated calls to the Scan function. Use the NewScanner() function
to create and initalize a scanner. In addition, convenience functions are
provided to initialize a scanner from a byte slice or a string.

This module was originally targeted at parsing non-trivial struct tags, but
it was then made more general so that it could be used for other purposes.

TokenScanner recognizes whitespace, comments (C++ style line and
multi-line), symbols (e.g., used for to denote operators), string literals
(quoted using one of " ' `), integers, and floats. By default, a
TokenScanner skips white space and comments, identifiers begin with a
letter and may contain letters or digits or "_" after the first character.
It may be customized to recognize additional quotes or define identifiers
differently, etc.

Example:


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

Output:


	nofile:3:5 (31)  - Ident  -> if
	nofile:3:8 (34)  - Ident  -> a
	nofile:3:10 (36) - Symbol -> >
	nofile:3:12 (38) - Int    -> 5
	nofile:3:14 (40) - Symbol -> {
	nofile:4:9 (50)  - Ident  -> b
	nofile:4:11 (52) - Symbol -> =
	nofile:4:13 (54) - String -> "this is a string"
	nofile:4:30 (72) - Symbol -> ;
	nofile:5:9 (82)  - Ident  -> c
	nofile:5:11 (84) - Symbol -> =
	nofile:5:13 (86) - Float  -> 7.2
	nofile:5:16 (89) - Symbol -> ;
	nofile:6:5 (95)  - Symbol -> }




## <a name="pkg-index">Index</a>
* [func IsDigitRune(ch rune, i int, runes []rune) bool](#IsDigitRune)
* [func IsEscapeRune(ch rune, i int, runes []rune) bool](#IsEscapeRune)
* [func IsIdentRune(ch rune, i int, runes []rune) bool](#IsIdentRune)
* [func IsQuoteRune(ch rune) (bool, rune)](#IsQuoteRune)
* [func IsQuoteRuneFancy(ch rune) (bool, rune)](#IsQuoteRuneFancy)
* [func IsSpaceRune(ch rune, i int, runes []rune) bool](#IsSpaceRune)
* [func IsSymbolRune(ch rune, i int, runes []rune) bool](#IsSymbolRune)
* [type Position](#Position)
  * [func (p *Position) String() string](#Position.String)
* [type Token](#Token)
  * [func (t *Token) String() string](#Token.String)
* [type TokenScanner](#TokenScanner)
  * [func NewScanner(r io.Reader) *TokenScanner](#NewScanner)
  * [func NewScannerBytes(b []byte) *TokenScanner](#NewScannerBytes)
  * [func NewScannerString(s string) *TokenScanner](#NewScannerString)
  * [func (ts *TokenScanner) Err() error](#TokenScanner.Err)
  * [func (ts *TokenScanner) Init(r io.Reader)](#TokenScanner.Init)
  * [func (ts *TokenScanner) Position() *Position](#TokenScanner.Position)
  * [func (ts *TokenScanner) Scan() bool](#TokenScanner.Scan)
  * [func (ts *TokenScanner) SetEOL(eol rune)](#TokenScanner.SetEOL)
  * [func (ts *TokenScanner) SetFilename(filename string)](#TokenScanner.SetFilename)
  * [func (ts *TokenScanner) Token() *Token](#TokenScanner.Token)
  * [func (ts *TokenScanner) TokenText() string](#TokenScanner.TokenText)
  * [func (ts *TokenScanner) TokenTextNoQuotes() string](#TokenScanner.TokenTextNoQuotes)
  * [func (ts *TokenScanner) UnreadToken() error](#TokenScanner.UnreadToken)
* [type TokenType](#TokenType)
  * [func (t TokenType) String() string](#TokenType.String)


#### <a name="pkg-files">Package files</a>
[classes.go](/src/github.com/cuberat/go-textparser/classes.go) [textparser.go](/src/github.com/cuberat/go-textparser/textparser.go) 





## <a name="IsDigitRune">func</a> [IsDigitRune](/src/target/classes.go?s=2604:2655#L60)
``` go
func IsDigitRune(ch rune, i int, runes []rune) bool
```
This function is the default value for the `IsDigitRune` field in
`TokenScanner`. Where `i` is the index of `ch` in the current token parse,
and `runes` is the list of runes already excepted for the current token.



## <a name="IsEscapeRune">func</a> [IsEscapeRune](/src/target/classes.go?s=1737:1791#L27)
``` go
func IsEscapeRune(ch rune, i int, runes []rune) bool
```
This function is the default value for the `IsEscapeRune` field in
`TokenScanner`. It returns true if the character is a reverse solidus (\).
Where `i` is the index of `ch` in the current token parse, and `runes` is
the list of runes already excepted for the current token.



## <a name="IsIdentRune">func</a> [IsIdentRune](/src/target/classes.go?s=2915:2966#L67)
``` go
func IsIdentRune(ch rune, i int, runes []rune) bool
```
This function is the default value for the `IsIdentRune` field in
`TokenScanner`. Where `i` is the index of `ch` in the current token parse,
and `runes` is the list of runes already excepted for the current token.



## <a name="IsQuoteRune">func</a> [IsQuoteRune](/src/target/classes.go?s=4354:4392#L126)
``` go
func IsQuoteRune(ch rune) (bool, rune)
```
This function is the default value for the `IsQuoteRune` field in
`TokenScanner`. It only treats single quotes ('), double quotes ("), and
back ticks (`) as quote characters.



## <a name="IsQuoteRuneFancy">func</a> [IsQuoteRuneFancy](/src/target/classes.go?s=3707:3750#L102)
``` go
func IsQuoteRuneFancy(ch rune) (bool, rune)
```
Alternative predicate for determing opening quote characters. Set the
`IsQuoteRune` field `TokenScanner` to this function to consider the
following to be quote characters (the plain quotes from English, as well as
more fancy ones specified in Unicode).


	"" - U+0022,U+0022
	'' - U+0027,U+0027
	“” - U+201C,U+201D
	‘’ - U+2018,U+2019
	‹› - U+2039,U+203A
	«» - U+00AB,U+00BB



## <a name="IsSpaceRune">func</a> [IsSpaceRune](/src/target/classes.go?s=4817:4868#L142)
``` go
func IsSpaceRune(ch rune, i int, runes []rune) bool
```
This function is the default value for the `IsIdentRune` field in
`TokenScanner`. Where `i` is the index of `ch` in the current token parse,
and `runes` is the list of runes already excepted for the current token.



## <a name="IsSymbolRune">func</a> [IsSymbolRune](/src/target/classes.go?s=2085:2137#L38)
``` go
func IsSymbolRune(ch rune, i int, runes []rune) bool
```
This function is the default value for the `IsSymbolRune` field in
`TokenScanner`. Where `i` is the index of `ch` in the current token parse,
and `runes` is the list of runes already excepted for the current token.




## <a name="Position">type</a> [Position](/src/target/textparser.go?s=4113:4351#L105)
``` go
type Position struct {
    Filename string // Filename, if any.
    Offset   int    // Byte offset (starting at 0).
    Line     int    // Line number (starting at 1).
    Column   int    // Column number (in characters, starting at 1).
}
```
Represents the position of the current token.










### <a name="Position.String">func</a> (\*Position) [String](/src/target/textparser.go?s=4413:4447#L113)
``` go
func (p *Position) String() string
```
Returns a string representation of the current position.




## <a name="Token">type</a> [Token](/src/target/textparser.go?s=4553:4822#L119)
``` go
type Token struct {
    Text      string    // The text of the token.
    NumBytes  int       // Number of bytes in the token.
    NumChars  int       // Number of characters/runes in the token.
    FirstRune rune      // First rune in the token.
    Type      TokenType // The type of token.
}
```
A Token.










### <a name="Token.String">func</a> (\*Token) [String](/src/target/textparser.go?s=4824:4855#L127)
``` go
func (t *Token) String() string
```



## <a name="TokenScanner">type</a> [TokenScanner](/src/target/textparser.go?s=5004:7610#L134)
``` go
type TokenScanner struct {

    // Indicator to skip whitespace tokens.
    SkipWhitespace bool

    // Indicator to skip comment tokens.
    SkipComments bool

    // The most recent Token generated by a call to Scan().
    LastToken *Token

    // Predicate controlling the characters accepted as the i'th rune in an
    // identifier (starting at zero). `runes` is the slice of runes accepted
    // so far for this token. The set of valid characters must not
    // intersect with the set of white space characters. The default is the
    // IsIdentRune function defined in this module.
    IsIdentRune func(ch rune, i int, runes []rune) bool

    // Predicate controlling the characters accepted as the i'th rune in a run
    // of white space. `runes` is the slice of runes accepted so far for this
    // token. The default value is `unicode.IsSpace()`, which decides based on
    // Unicode's White Space property.
    IsSpaceRune func(ch rune, i int, runes []rune) bool

    // Predicate controlling the characters accepted as quoting runes. Returns
    // true/false, as well as the corresponding closing quote rune. The
    // default is the IsQuoteRune define in this module.
    IsQuoteRune func(ch rune) (bool, rune)

    // Predicate controlling the characters accepted as escape runes, e.g.,
    // for escaping a quote character inside quotes. `i` is the index of the
    // current rune being considered for this token. `runes` is the list of
    // runes already accepted for this token.
    IsEscapeRune func(ch rune, i int, runes []rune) bool

    // Predicate controlling the characters accepted as the i'th rune in a
    // symbol token (starting at zero). `runes` is the list of runes already
    // accepted for this token. The default predicate considers each symbol to
    // be its own token. This can be customized to allow for tokens to consist
    // of groups of certain symbols. One of the examples in the documentation
    // and test file does this.
    IsSymbolRune func(ch rune, i int, runes []rune) bool

    // Predicate controlling the characters accepted as numeric digits. `i` is
    // the index of the current rune being considered for this token. `runes`
    // is the list of runes already accepted for this token.
    IsDigitRune func(ch rune, i int, runes []rune) bool
    // contains filtered or unexported fields
}
```
A TokenScanner.







### <a name="NewScanner">func</a> [NewScanner](/src/target/textparser.go?s=7859:7901#L204)
``` go
func NewScanner(r io.Reader) *TokenScanner
```
Returns a new TokenScanner initialized with the provided reader.


### <a name="NewScannerBytes">func</a> [NewScannerBytes](/src/target/textparser.go?s=8229:8273#L218)
``` go
func NewScannerBytes(b []byte) *TokenScanner
```
Returns a TokenScanner initialized with the contents of the provided
byte slice.


### <a name="NewScannerString">func</a> [NewScannerString](/src/target/textparser.go?s=8047:8092#L212)
``` go
func NewScannerString(s string) *TokenScanner
```
Returns a TokenScanner initialized with the contents of the provided
string.





### <a name="TokenScanner.Err">func</a> (\*TokenScanner) [Err](/src/target/textparser.go?s=9077:9112#L252)
``` go
func (ts *TokenScanner) Err() error
```
Returns the last error encountered.




### <a name="TokenScanner.Init">func</a> (\*TokenScanner) [Init](/src/target/textparser.go?s=8467:8508#L224)
``` go
func (ts *TokenScanner) Init(r io.Reader)
```
Initializes a TokenScanner with the provided reader. This is only needed if
a TokenScanner is created outside of one of the New* functions.




### <a name="TokenScanner.Position">func</a> (\*TokenScanner) [Position](/src/target/textparser.go?s=7723:7767#L199)
``` go
func (ts *TokenScanner) Position() *Position
```
Returns position information for the current state. The same Position
object is used throughout parsing.




### <a name="TokenScanner.Scan">func</a> (\*TokenScanner) [Scan](/src/target/textparser.go?s=11487:11522#L339)
``` go
func (ts *TokenScanner) Scan() bool
```
Scans the next token, skipping whitespace and comments, unless configured
differently. Returns true if another token was found. Returns false when
parsing is completed. Check ts.Err() for parsing errors.




### <a name="TokenScanner.SetEOL">func</a> (\*TokenScanner) [SetEOL](/src/target/textparser.go?s=10604:10644#L309)
``` go
func (ts *TokenScanner) SetEOL(eol rune)
```
Sets the rune considered to be the end-of-line character.




### <a name="TokenScanner.SetFilename">func</a> (\*TokenScanner) [SetFilename](/src/target/textparser.go?s=10722:10774#L314)
``` go
func (ts *TokenScanner) SetFilename(filename string)
```
Sets the file name returned in the Position object.




### <a name="TokenScanner.Token">func</a> (\*TokenScanner) [Token](/src/target/textparser.go?s=9205:9243#L257)
``` go
func (ts *TokenScanner) Token() *Token
```
Returns the most recent token generated by a call to Scan().




### <a name="TokenScanner.TokenText">func</a> (\*TokenScanner) [TokenText](/src/target/textparser.go?s=10007:10049#L286)
``` go
func (ts *TokenScanner) TokenText() string
```
Returns the text from the most recent token generated by a call to Scan().




### <a name="TokenScanner.TokenTextNoQuotes">func</a> (\*TokenScanner) [TokenTextNoQuotes](/src/target/textparser.go?s=10288:10338#L296)
``` go
func (ts *TokenScanner) TokenTextNoQuotes() string
```
Returns the text from the most recent token generated by a call to Scan().
If the token is a quoted string, the surrounding quotes are removed.




### <a name="TokenScanner.UnreadToken">func</a> (\*TokenScanner) [UnreadToken](/src/target/textparser.go?s=9582:9625#L269)
``` go
func (ts *TokenScanner) UnreadToken() error
```
Pretends the current token was not read. The next call to `Scan()` and
`Token()` will return the current token. Once invoked, further
`UnreadToken()` calls are invalid until the next `Scan()` call.




## <a name="TokenType">type</a> [TokenType](/src/target/textparser.go?s=3575:3593#L80)
``` go
type TokenType int
```

``` go
const (
    TokenTypeWhitespace TokenType = iota
    TokenTypeIdent
    TokenTypeString
    TokenTypeComment
    TokenTypeInt
    TokenTypeFloat
    TokenTypeSymbol
)
```
Supported token types.










### <a name="TokenType.String">func</a> (TokenType) [String](/src/target/textparser.go?s=3843:3877#L94)
``` go
func (t TokenType) String() string
```
Returns a string representation of the token type.








- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
