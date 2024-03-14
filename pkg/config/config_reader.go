package config

// TODO: separate token.go, token_reader.go, config_reader.go

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"
)

const EOF Token = ""

var whitespace = []Token{" ", "\t", "\n", "\r"}
var special = []Token{"{", "}", ":"}

// ----------------------------------------
type Token string

func (t Token) String() string {
	return string(t)
}

func (t Token) Quote() string {
	if t == "" {
		return "EOF"
	} else {
		return fmt.Sprintf("'%s'", t)
	}
}

func (t Token) IsSpecial() bool {
	return slices.Contains(special, t)
}

func (t Token) IsLiteral() bool {
	return t != EOF && !t.IsSpecial()
}

// ----------------------------------------
type tokenReader struct {
	tokens []Token
	err    error
}

func newTokenReader(r io.Reader) (reader *tokenReader) {
	// TODO: collect tokens on the run
	tokens, err := collectTokens(r)
	reader = &tokenReader{tokens, err}
	return
}

func (r *tokenReader) next() (t Token, err error) {
	if r.err != nil {
		err = r.err
		r.err = nil
		return
	}
	if len(r.tokens) < 1 {
		return EOF, nil
	}
	t = r.tokens[0]
	r.tokens = r.tokens[1:]
	return
}

func collectTokens(r io.Reader) (tokens []Token, err error) {
	var reader = bufio.NewReader(r)
	var tok strings.Builder

	for {
		var r, _, err = reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				if tok.Len() > 0 {
					tokens = append(tokens, Token(tok.String()))
					tok.Reset()
				}
				break
			} else {
				return nil, err
			}
		}

		// whitespace ends any token that was being accumulated
		if slices.Contains(whitespace, Token(r)) {
			if tok.Len() > 0 {
				tokens = append(tokens, Token(tok.String()))
				tok.Reset()
			}
			continue
		}

		// check for special characters
		if slices.Contains(special, Token(r)) {
			if tok.Len() > 0 {
				tokens = append(tokens, Token(tok.String()))
				tok.Reset()
			}
			tok.WriteRune(r)
			tokens = append(tokens, Token(tok.String()))
			tok.Reset()
			continue
		}

		// build normal token
		tok.WriteRune(r)
	}

	return
}

// ----------------------------------------
type Reader struct {
	tokens *tokenReader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{newTokenReader(r)}
}

// Read next token. If reader reached EOF, return ""
func (r *Reader) ReadNext() (Token, error) {
	return r.tokens.next()
}

// Read next token and check that it matches exp
func (r *Reader) ReadExact(exp Token) error {
	var token, err = r.ReadNext()
	if err != nil {
		return err
	}
	if token != exp {
		return ErrUnexpectedToken(token, exp.Quote())
	}
	return nil
}

// Read next token and check that it is a literal (not special or EOF)
func (r *Reader) ReadLiteral() (t Token, err error) {
	t, err = r.ReadNext()
	if err != nil {
		return
	}
	if !t.IsLiteral() {
		t = EOF
		err = ErrUnexpectedToken(t, "a valid name")
		return
	}
	return
}

// Check that there is a ":" and read next literal token
func (r *Reader) ReadProperty() (t Token, err error) {
	if err = r.ReadExact(":"); err != nil {
		return
	}
	return r.ReadLiteral()
}

// Same as ReadProperty(), but gets the token string
func (r *Reader) ReadPropertyName() (s string, err error) {
	var t Token
	t, err = r.ReadProperty()
	if err != nil {
		return
	}
	s = t.String()
	return
}

// Read a structure block, each literal token is passed to parseField function.
//
// Example:
//
//	{
//	  field_a [rest processed by parseField]
//	  field_b [...]
//	}
func (r *Reader) ReadStruct(parseField func(tokens *Reader, field Token) error) (err error) {
	if err = r.ReadExact("{"); err != nil {
		return
	}
	for {
		var token Token
		token, err = r.ReadNext()
		if err != nil {
			return
		}
		if token == "}" {
			break
		} else if !token.IsLiteral() {
			return ErrUnexpectedToken(token, "property name or '}'")
		}

		err = parseField(r, token)
		if err != nil {
			return
		}
	}
	return
}

func ErrUnexpectedToken(t Token, expect string) error {
	return fmt.Errorf("unexpected %s, %s was expected", t.Quote(), expect)
}
func ErrUnrecognized(t Token, exp string) error {
	return fmt.Errorf("%s is not a recognized%s", t.Quote(), exp)
}
