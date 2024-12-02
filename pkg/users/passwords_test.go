package users

import (
	"errors"
	"testing"
)

func TestPasswordFull(t *testing.T) {
	tests := []struct {
		name        string
		passwordIn  string
		passwordCmp string
		wantErr     error
	}{
		{
			name:        "Match",
			passwordIn:  "password",
			passwordCmp: "password",
			wantErr:     nil,
		}, {
			name:        "Mismatch",
			passwordIn:  "password",
			passwordCmp: "password2",
			wantErr:     ErrPasswordMismatch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := createEncodedPassword(tt.passwordIn)
			err := checkPassword(encoded, tt.passwordCmp)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("checkPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPasswordMatch(t *testing.T) {
	tests := []struct {
		name     string
		encoded  string
		password string
		wantErr  error
	}{
		{
			name:     "Match",
			encoded:  "1zB7mniBv/M00n4Xe7Cs2fP+t5mJIqQs7wbbRw1Djhfo66f5bc90dcf934ea",
			password: "password",
			wantErr:  nil,
		}, {
			name:     "Mismatch",
			encoded:  "1zB7mniBv/M00n4Xe7Cs2fP+t5mJIqQs7wbbRw1Djhfo66f5bc90dcf934ea",
			password: "password2",
			wantErr:  ErrPasswordMismatch,
		}, {
			name:     "ShortHash",
			encoded:  "1zB7mniBv/M00n4Xe7Cs2fP+t5mJIqQs7wbbRw1Djhfo66f5bc90dcf934e",
			password: "password",
			wantErr:  ErrPasswordEncoding,
		}, {
			name:     "BadHash",
			encoded:  "1zB7mniBv/M00n4Xe()s2fP+t5mJIqQs7wbbRw1Djhfo66f5bc90dcf934ea",
			password: "password",
			wantErr:  ErrPasswordEncoding,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkPassword(tt.encoded, tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("checkPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifyEncodedPassword(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
		wantErr error
	}{
		{
			name:    "Match",
			encoded: "1zB7mniBv/M00n4Xe7Cs2fP+t5mJIqQs7wbbRw1Djhfo66f5bc90dcf934ea",
			wantErr: nil,
		}, {
			name:    "ShortHash",
			encoded: "1zB7mniBv/M00n4Xe7Cs2fP+t5mJIqQs7wbbRw1Djhfo66f5bc90dcf934e",
			wantErr: ErrPasswordEncoding,
		}, {
			name:    "BadHash",
			encoded: "1zB7mniBv/M00n4Xe()s2fP+t5mJIqQs7wbbRw1Djhfo66f5bc90dcf934ea",
			wantErr: ErrPasswordEncoding,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyEncodedPassword(tt.encoded)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("checkPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
