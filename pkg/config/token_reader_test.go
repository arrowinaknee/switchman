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
			name:  "Literals",
			input: "test case",
			want:  []Token{"test", "case", EOF},
		}, {
			name:  "Special",
			input: "test:{case }",
			want:  []Token{"test", ":", "{", "case", "}", EOF},
		}, {
			name:  "Empty",
			input: "",
			want:  []Token{EOF},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokens []Token
			var reader = newTokenReader(strings.NewReader(tt.input))
			for {
				tok, err := reader.next()
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
