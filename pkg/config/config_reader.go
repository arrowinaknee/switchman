package main

// TODO: separate token.go, token_reader.go, config_reader.go

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"
)

const EOF token = ""

var whitespace = []token{" ", "\t", "\n", "\r"}
var special = []token{"{", "}", ":"}

// ----------------------------------------
type token string

func (t token) String() string {
	return string(t)
}

func (t token) Quote() string {
	if t == "" {
		return "EOF"
	} else {
		return fmt.Sprintf("'%s'", t)
	}
}

func (t token) IsSpecial() bool {
	return slices.Contains(special, t)
}

func (t token) IsLiteral() bool {
	return t != EOF && !t.IsSpecial()
}

// ----------------------------------------
type tokenReader struct {
	tokens []token
	err    error
}

func newTokenReader(r io.Reader) (reader *tokenReader) {
	// TODO: collect tokens on the run
	tokens, err := collectTokens(r)
	reader = &tokenReader{tokens, err}
	return
}

func (r *tokenReader) next() (t token, err error) {
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

func collectTokens(r io.Reader) (tokens []token, err error) {
	var reader = bufio.NewReader(r)
	var tok strings.Builder

	for {
		var r, _, err = reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				if tok.Len() > 0 {
					tokens = append(tokens, token(tok.String()))
					tok.Reset()
				}
				break
			} else {
				return nil, err
			}
		}

		// whitespace ends any token that was being accumulated
		if slices.Contains(whitespace, token(r)) {
			if tok.Len() > 0 {
				tokens = append(tokens, token(tok.String()))
				tok.Reset()
			}
			continue
		}

		// check for special characters
		if slices.Contains(special, token(r)) {
			if tok.Len() > 0 {
				tokens = append(tokens, token(tok.String()))
				tok.Reset()
			}
			tok.WriteRune(r)
			tokens = append(tokens, token(tok.String()))
			tok.Reset()
			continue
		}

		// build normal token
		tok.WriteRune(r)
	}

	return
}

// ----------------------------------------
type ConfigReader struct {
	tokens *tokenReader
}

func NewConfigReader(r io.Reader) *ConfigReader {
	return &ConfigReader{newTokenReader(r)}
}

// Read next token. If reader reached EOF, return ""
func (r *ConfigReader) ReadNext() (token, error) {
	return r.tokens.next()
}

// Read next token and check that it matches exp
func (r *ConfigReader) ReadExact(exp token) error {
	var token, err = r.ReadNext()
	if err != nil {
		return err
	}
	if token != exp {
		return errUnexpectedToken(token, exp.Quote())
	}
	return nil
}

// Read next token and check that it is a literal (not special or EOF)
func (r *ConfigReader) ReadLiteral() (t token, err error) {
	t, err = r.ReadNext()
	if err != nil {
		return
	}
	if !t.IsLiteral() {
		t = EOF
		err = errUnexpectedToken(t, "a valid name")
		return
	}
	return
}

// Check that there is a ":" and read next literal token
func (r *ConfigReader) ReadProperty() (t token, err error) {
	if err = r.ReadExact(":"); err != nil {
		return
	}
	return r.ReadLiteral()
}

// Same as ReadProperty(), but gets the token string
func (r *ConfigReader) ReadPropertyName() (s string, err error) {
	var t token
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
func (r *ConfigReader) ReadStruct(parseField func(tokens *ConfigReader, field token) error) (err error) {
	if err = r.ReadExact("{"); err != nil {
		return
	}
	for {
		var token token
		token, err = r.ReadNext()
		if err != nil {
			return
		}
		if token == "}" {
			break
		} else if !token.IsLiteral() {
			return errUnexpectedToken(token, "property name or '}'")
		}

		err = parseField(r, token)
		if err != nil {
			return
		}
	}
	return
}

func errUnexpectedToken(t token, expect string) error {
	return fmt.Errorf("unexpected %s, %s was expected", t.Quote(), expect)
}
func errUnrecognized(t token, exp string) error {
	return fmt.Errorf("%s is not a recognized%s", t.Quote(), exp)
}
