package config

import (
	"fmt"
	"io"
)

type Reader struct {
	tokens   *tokenReader
	curToken Token
	tokenPos TokenPosition
}

func NewReader(r io.Reader) *Reader {
	return &Reader{tokens: newTokenReader(r)}
}

// Read next token. If reader reached EOF, return ""
func (r *Reader) ReadNext() (Token, error) {
	var err error
	r.curToken, r.tokenPos, err = r.tokens.next()
	return r.curToken, err
}

// Read next token and check that it matches exp
func (r *Reader) ReadExact(exp Token) error {
	var token, err = r.ReadNext()
	if err != nil {
		return err
	}
	if token != exp {
		return r.ErrUnexpectedToken(exp.Quote())
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
		err = r.ErrUnexpectedToken("a valid literal")
		return
	}
	return
}

func (r *Reader) ReadSeparator() error {
	return r.ReadExact(":")
}

func (r *Reader) ReadName() (t Token, err error) {
	t, err = r.ReadLiteral()
	if err != nil {
		return
	}
	if !t.IsName() {
		err = r.ErrUnexpectedToken("a valid name")
		return
	}
	return
}

func (r *Reader) ReadString() (t Token, err error) {
	t, err = r.ReadLiteral()
	if err != nil {
		return
	}
	t, err = t.Unescaped()
	if err != nil {
		return
	}
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
			return r.ErrUnexpectedToken("property name or '}'")
		}

		err = parseField(r, token)
		if err != nil {
			return
		}
	}
	return
}

func (r *Reader) ErrUnexpectedToken(expect string) error {
	return fmt.Errorf("%d:%d: %s was expected, got %s", r.tokenPos.Line, r.tokenPos.Col, expect, r.curToken.Quote())
}
func (r *Reader) ErrUnrecognized(exp string) error {
	return fmt.Errorf("%d:%d: %s is not a recognized %s", r.tokenPos.Line, r.tokenPos.Col, r.curToken.Quote(), exp)
}
func (r *Reader) ErrInvalid(exp string) error {
	return fmt.Errorf("%d:%d: %s is not a valid %s", r.tokenPos.Line, r.tokenPos.Col, r.curToken.Quote(), exp)
}
