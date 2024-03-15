package config

import (
	"fmt"
	"io"
)

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
