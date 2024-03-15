package config

import (
	"fmt"
	"slices"
)

const EOF Token = ""

var whitespace = []Token{" ", "\t", "\n", "\r"}
var special = []Token{"{", "}", ":"}

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
