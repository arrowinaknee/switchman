package config

import (
	"reflect"
	"strings"
	"testing"
)

// Read all tokens in input with tokenReader.next(), then check results
func TestTokenReader(t *testing.T) {
	type testCase struct {
		name  string
		input string
		want  []Token
	}
	var tests = []testCase{
		{
			name:  "literals",
			input: "test case",
			want:  []Token{"test", "case", EOF},
		}, {
			name:  "special",
			input: "test:{case \t}",
			want:  []Token{"test", ":", "{", "case", "}", EOF},
		}, {
			name:  "empty",
			input: "",
			want:  []Token{EOF},
		}, {
			name: "comment",
			input: `test {
				case # test case
				# comment {:}
			}`,
			want: []Token{"test", "{", "case", "}", EOF},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokens []Token
			var reader = newTokenReader(strings.NewReader(tt.input))
			for {
				tok, _, err := reader.next()
				if err != nil {
					t.Errorf("tokenReader.next() error = %v", err)
					return
				}
				tokens = append(tokens, tok)
				if tok == EOF {
					break
				}
			}
			if !reflect.DeepEqual(tokens, tt.want) {
				t.Errorf("tokenReader.next() tokens = %v, want = %v", tokens, tt.want)
			}
		})
	}
}

func TestTokenPosition(t *testing.T) {
	type tokenWithPos struct {
		t Token
		p TokenPosition
	}
	input := `test {
case :
}`
	want := []tokenWithPos{
		{"test", *position(1, 1)},
		{"{", *position(1, 6)},
		{"case", *position(2, 1)},
		{":", *position(2, 6)},
		{"}", *position(3, 1)},
		{EOF, *position(3, 2)},
	}
	var got []tokenWithPos
	var reader = newTokenReader(strings.NewReader(input))
	for {
		tok, pos, err := reader.next()
		if err != nil {
			t.Errorf("tokenReader.next() error = %v", err)
			return
		}
		got = append(got, tokenWithPos{tok, pos})
		if tok == EOF {
			break
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tokenReader.next() tokens = %v, want = %v", got, want)
	}
}
