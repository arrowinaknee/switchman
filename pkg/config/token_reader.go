package config

import (
	"bufio"
	"io"
	"slices"
	"strings"
)

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
