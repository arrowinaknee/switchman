package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
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
}

func NewTokenReader(r io.Reader) (reader *tokenReader, err error) {
	// TODO: collect tokens on the run from ReadToken
	tokens, err := collectTokens(r)
	if err != nil {
		return nil, err
	}
	reader = &tokenReader{
		tokens: tokens,
	}
	return
}

// Read next token. If reader reached EOF, return ""
func (r *tokenReader) ReadNext() token {
	if len(r.tokens) < 1 {
		return EOF
	}
	var token = r.tokens[0]
	r.tokens = r.tokens[1:]
	return token
}

// Read next token and check that it matches exp
func (r *tokenReader) ReadExact(exp token) {
	var token = r.ReadNext()
	if token != exp {
		log.Fatalf("Unexpected %s, %s expected", token.Quote(), exp.Quote())
	}
}

func (r *tokenReader) ReadLiteral() token {
	var token = r.ReadNext()
	if !token.IsLiteral() {
		log.Fatalf("Unexpected %s, a valid name expected", token.Quote())
	}
	return token
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
