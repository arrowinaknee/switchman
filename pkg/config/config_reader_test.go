package config

import (
	"reflect"
	"strings"
	"testing"
)

func TestReader_ReadNext(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Token
		wantErr bool
	}{
		{
			name:  "eof",
			input: "   ",
			want:  EOF,
		}, {
			name:  "literal",
			input: "\n literal ",
			want:  "literal",
		}, {
			name:  "special",
			input: ":literal",
			want:  ":",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			got, err := r.ReadNext()
			if (err != nil) != tt.wantErr {
				t.Errorf("Reader.ReadNext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Reader.ReadNext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReader_ReadExact(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		exp     Token
		wantErr bool
	}{
		{
			name:    "eof_expected",
			input:   "",
			exp:     EOF,
			wantErr: false,
		}, {
			name:    "eof_unexpected",
			input:   "",
			exp:     "literal",
			wantErr: true,
		}, {
			name:    "literal_correct",
			input:   "  literal{}",
			exp:     "literal",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			if err := r.ReadExact(tt.exp); (err != nil) != tt.wantErr {
				t.Errorf("Reader.ReadExact() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReader_ReadLiteral(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Token
		wantErr bool
	}{
		{
			name:  "correct",
			input: "literal",
			want:  "literal",
		}, {
			name:    "special",
			input:   ": literal",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			got, err := r.ReadLiteral()
			if (err != nil) != tt.wantErr {
				t.Errorf("Reader.ReadLiteral() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Reader.ReadLiteral() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReader_ReadSeparator(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"correct", ":", false},
		{"incorrect", "}", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			if err := r.ReadSeparator(); (err != nil) != tt.wantErr {
				t.Errorf("Reader.ReadSeparator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReader_ReadName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Token
		wantErr bool
	}{
		{"name", "camelCase_95", "camelCase_95", false},
		{"string", "'string'", "", true},
		{"special", "}", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			got, err := r.ReadName()
			if (err != nil) != tt.wantErr {
				t.Errorf("Reader.ReadName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reader.ReadName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReader_ReadString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Token
		wantErr bool
	}{
		{"open", "string another", "string", false},
		{"quotes_single", "'two words'", "two words", false},
		{"quotes_double", `"two words"`, "two words", false},
		{"not_terminated", "'string\n", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			got, err := r.ReadString()
			if (err != nil) != tt.wantErr {
				t.Errorf("Reader.ReadString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reader.ReadString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReader_ReadStruct(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantFields []Token
		wantErr    bool
	}{
		{
			name: "normal",
			input: `{
				field1
				field2
			}`,
			wantFields: []Token{"field1", "field2"},
		}, {
			name: "empty",
			input: `{
			}`,
			wantFields: nil,
		}, {
			name: "not_open",
			input: `field1
			field2`,
			wantErr: true,
		}, {
			name: "incomplete",
			input: `{
				field1
				field2`,
			wantErr: true,
		}, {
			name: "non_literal",
			input: `{
				:
				field
			}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			var gotFields []Token
			parseField := func(tokens *Reader, field Token) error {
				gotFields = append(gotFields, field)
				return nil
			}
			err := r.ReadStruct(parseField)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reader.ReadStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(gotFields, tt.wantFields) {
				t.Errorf("Reader.ReadStruct() parseField called for fields %v, want %v", gotFields, tt.wantFields)
			}
		})
	}
}
