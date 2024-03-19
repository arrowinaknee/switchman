package config

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

const EOF Token = ""

var special = []Token{"{", "}", ":"}
var name_regexp = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

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

// IsName checks wether token can be used as a name.
//
// Name is a literal with at least one character and containing only english alphanumerical characters or underlines '_'.
func (t Token) IsName() bool {
	if !t.IsLiteral() {
		return false
	}
	return name_regexp.MatchString(t.String())
}

func (t Token) Unescaped() (Token, error) {
	if len(t) > 0 {
		// string in quotes
		if quote := t[0]; quote == '"' || quote == '\'' {
			// string is terminated and the closing quote is not escaped
			if len(t) < 2 || t[len(t)-1] != quote || t[len(t)-2] == '\\' {
				return EOF, fmt.Errorf("quoted string literal not terminated")
			}
			// remove quotes
			t = t[1 : len(t)-1]
			// unescape quotes
			t = Token(strings.ReplaceAll(t.String(), fmt.Sprintf("\\%c", quote), string(quote)))
		}
	}
	// error any other escapes
	if i := strings.IndexByte(t.String(), '\\'); i != -1 {
		if i < len(t)-1 {
			return EOF, fmt.Errorf("%s is not a recognized escape sequence", t[i:i+2])
		} else {
			return EOF, fmt.Errorf("escape sequence incomplete")
		}
	}
	return t, nil
}
