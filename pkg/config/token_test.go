package config

import (
	"testing"
)

func TestToken_IsName(t *testing.T) {
	tests := []struct {
		name string
		tok  Token
		want bool
	}{
		{"simple", "name", true},
		{"advanced", "_camelCase_95", true},
		{"special", "}", false},
		{"escaped", `name\:`, false},
		{"spaces", "with space", false},
		{"quotes", "'quoted'", false},
		{"EOF", EOF, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tok.IsName(); got != tt.want {
				t.Errorf("Token.IsName(%s) = %v, want %v", tt.tok.Quote(), got, tt.want)
			}
		})
	}
}

func TestToken_Unescape(t *testing.T) {
	tests := []struct {
		name    string
		tok     Token
		want    string
		wantErr bool
	}{
		{"not_quoted", "string", "string", false},
		{"quoted", "'string'", "string", false},
		{"not_terminated", "'string", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tok.Unescape()
			if (err != nil) != tt.wantErr {
				t.Errorf("Token.Unescape() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got != tt.want {
				t.Errorf("Token.Unescape() = %v, want %v", got, tt.want)
			}
		})
	}
}
