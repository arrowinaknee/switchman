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
			input:   ": litera;",
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

func TestReader_ReadProperty(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Token
		wantErr bool
	}{
		{
			name:  "correct",
			input: ": value",
			want:  "value",
		}, {
			name:    "no_colon",
			input:   " value",
			wantErr: true,
		}, {
			name:    "no_value",
			input:   ": }",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			got, err := r.ReadProperty()
			if (err != nil) != tt.wantErr {
				t.Errorf("Reader.ReadProperty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Reader.ReadProperty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReader_ReadPropertyName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "correct",
			input: ": value",
			want:  "value",
		}, {
			name:    "no_colon",
			input:   " value",
			wantErr: true,
		}, {
			name:    "no_value",
			input:   ": }",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			got, err := r.ReadPropertyName()
			if (err != nil) != tt.wantErr {
				t.Errorf("Reader.ReadPropertyName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Reader.ReadPropertyName() = %v, want %v", got, tt.want)
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
