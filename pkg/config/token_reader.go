package config

import (
	"bufio"
	"io"
	"slices"
	"strings"
)

type tokenReader struct {
	reader *bufio.Reader
	token  strings.Builder
}

func newTokenReader(r io.Reader) (reader *tokenReader) {
	return &tokenReader{
		reader: bufio.NewReader(r),
	}
}

func (r *tokenReader) popToken() (t Token) {
	t = Token(r.token.String())
	r.token.Reset()
	return
}

func (r *tokenReader) next() (t Token, err error) {
	var whitespace = []rune{' ', '\t', '\n', '\r'}

	for {
		var c rune
		c, _, err = r.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				if r.token.Len() > 0 {
					t = r.popToken()
					return t, nil
				}
				return EOF, nil
			}
			return
		}

		if slices.Contains(whitespace, c) {
			if r.token.Len() > 0 {
				t = r.popToken()
				return
			}
			continue
		}

		if Token(c).IsSpecial() {
			if r.token.Len() > 0 {
				t = r.popToken()
				r.token.WriteRune(c)
			} else {
				t = Token(c)
			}
			return
		}

		if Token(r.token.String()).IsSpecial() {
			t = r.popToken()
			r.token.WriteRune(c)
			return
		}
		r.token.WriteRune(c)
	}
}
