package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"
)

const EOF = ""

var whitespace = []string{" ", "\t", "\n", "\r"}
var special = []string{"{", "}", ":"}

type tokenReader struct {
	tokens []string
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

// Reads next token from ConfigReader. If reader reached EOF, returns ""
func (r *tokenReader) ReadToken() string {
	if len(r.tokens) < 1 {
		return EOF
	}
	var token = r.tokens[0]
	r.tokens = r.tokens[1:]
	return token
}

func (r *tokenReader) ReadExactToken(exp string) {
	var token = r.ReadToken()
	if token != exp {
		log.Fatalf("Unexpected %s, %s expected", TokenName(token), TokenName(exp))
	}
}

func collectTokens(r io.Reader) (tokens []string, err error) {
	var reader = bufio.NewReader(r)
	var tok strings.Builder

	for {
		var r, _, err = reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				if tok.Len() > 0 {
					tokens = append(tokens, tok.String())
					tok.Reset()
				}
				break
			} else {
				return nil, err
			}
		}

		// whitespace ends any token that was being accumulated
		if slices.Contains(whitespace, string(r)) {
			if tok.Len() > 0 {
				tokens = append(tokens, tok.String())
				tok.Reset()
			}
			continue
		}

		// check for special characters
		if slices.Contains(special, string(r)) {
			if tok.Len() > 0 {
				tokens = append(tokens, tok.String())
				tok.Reset()
			}
			tok.WriteRune(r)
			tokens = append(tokens, tok.String())
			tok.Reset()
			continue
		}

		// build normal token
		tok.WriteRune(r)
	}

	return
}

func TokenName(t string) string {
	if t == "" {
		return "EOF"
	} else {
		return fmt.Sprintf("'%s'", t)
	}
}
