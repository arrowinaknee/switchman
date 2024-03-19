package config

import (
	"fmt"
	"regexp"
	"slices"
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

func (t Token) Unescape() (s string, err error) {
	s = t.String()
	if len(s) > 0 {
		// string in quotes
		if quote := s[0]; quote == '"' || quote == '\'' {
			if len(s) < 2 || s[len(s)-1] != quote {
				err = fmt.Errorf("quoted string literal not terminated")
				return
			}
			// remove quotes
			s = s[1 : len(s)-1]
		}
	}
	return
}
