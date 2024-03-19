package config

import (
	"bufio"
	"io"
	"slices"
	"strings"
)

type TokenPosition struct {
	Line int
	Col  int
}

func startPosition() *TokenPosition {
	return &TokenPosition{
		Line: 1,
		Col:  0,
	}
}

func position(line, col int) *TokenPosition {
	return &TokenPosition{line, col}
}

func (t *TokenPosition) nextChar() {
	t.Col += 1
}

func (t *TokenPosition) nextLine() {
	t.Line += 1
	t.Col = 0
}

type tokenReader struct {
	reader   *bufio.Reader
	token    strings.Builder
	tokStart TokenPosition
	curPos   *TokenPosition
}

func newTokenReader(r io.Reader) (reader *tokenReader) {
	return &tokenReader{
		reader:   bufio.NewReader(r),
		tokStart: *startPosition(),
		curPos:   startPosition(),
	}
}

func (r *tokenReader) popToken() (t Token, p TokenPosition) {
	t = Token(r.token.String())
	r.token.Reset()
	p = r.tokStart
	return
}

func (r *tokenReader) writeRune(c rune) {
	if r.token.Len() == 0 {
		r.tokStart = *r.curPos
	}
	r.token.WriteRune(c)
}

func (r *tokenReader) firstChar() rune {
	if r.token.Len() == 0 {
		return rune(0)
	}
	return rune(r.token.String()[0])
}

func (r *tokenReader) next() (t Token, p TokenPosition, err error) {
	var whitespace = []rune{' ', '\t', '\n', '\r'}

	for {
		var c rune
		c, _, err = r.reader.ReadRune()
		r.curPos.nextChar()
		if err != nil {
			if err == io.EOF {
				err = nil
				if r.token.Len() > 0 {
					t, p = r.popToken()
					return
				} else {
					t, p = EOF, *r.curPos
				}
				return
			}
			return
		}

		if quote := r.firstChar(); quote == '"' || quote == '\'' {
			if c == quote {
				r.writeRune(c)
				t, p = r.popToken()
				return
			}
			// newlines can't be part of a string, but the error will be raised from token validation
			if c == '\n' {
				t, p = r.popToken()
				return
			}
			r.writeRune(c)
			continue
		}
		if c == '"' || c == '\'' {
			if r.token.Len() > 0 {
				t, p = r.popToken()
				r.writeRune(c)
				return
			}
			r.writeRune(c)
			continue
		}

		if slices.Contains(whitespace, c) {
			if c == '\n' {
				r.curPos.nextLine()
			}
			if r.token.Len() > 0 {
				t, p = r.popToken()
				return
			}
			continue
		}

		if c == '#' {
			// seek end of line
			err = r.processComment()
			if err != nil {
				return
			}
			// return the token that was before the comment, otherwise continue
			if r.token.Len() > 0 {
				t, p = r.popToken()
				return
			}
			continue
		}

		if Token(c).IsSpecial() {
			if r.token.Len() > 0 {
				t, p = r.popToken()
				r.writeRune(c)
			} else {
				t, p = Token(c), *r.curPos
			}
			return
		}

		if Token(r.token.String()).IsSpecial() {
			t, p = r.popToken()
			r.writeRune(c)
			return
		}
		r.writeRune(c)
	}
}

func (r *tokenReader) processComment() error {
	for {
		c, _, err := r.reader.ReadRune()
		r.curPos.nextChar()
		if err != nil {
			// EOF will be processed in the next iteration, any other error means the parsing failed
			if err == io.EOF {
				return nil
			}
			return err
		}
		if c == '\n' {
			r.curPos.nextLine()
			return nil
		}
	}
}
