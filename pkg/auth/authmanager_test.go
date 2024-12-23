package auth

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/arrowinaknee/switchman/pkg/settings"
	"github.com/golang-jwt/jwt/v5"
)

var testSecret = sha256.Sum256([]byte("testbytes"))
var badSecret = sha256.Sum256([]byte("badbytes"))

func TestTokenFull(t *testing.T) {
	auth := makeTestAuth()
	userid := "12345678"
	tok, err := auth.IssueToken(userid)
	if err != nil {
		t.Fatal(err)
	}
	gotid, err := auth.ProcessToken(tok)
	if err != nil {
		t.Fatal(err)
	}
	if gotid != userid {
		t.Errorf("expected user %q from token, got %q", userid, gotid)
	}
}

func TestIssueToken(t *testing.T) {
	tests := []struct {
		name    string
		userid  string
		wantErr bool
	}{
		{
			name:    "normal",
			userid:  "12345678",
			wantErr: false,
		}, {
			name:    "empty",
			userid:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := makeTestAuth()
			tokenStr, err := auth.IssueToken(tt.userid)
			if (err != nil) != tt.wantErr {
				t.Errorf("IssueToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				return testSecret[:], nil
			})
			if err != nil {
				t.Errorf("parse token: %v", err)
			}
			claims := token.Claims.(jwt.MapClaims)
			if claims["sub"] != tt.userid {
				t.Errorf("claims[\"sub\"] = %v, want %v", claims["sub"], tt.userid)
			}
			if claims["iss"] != "switchman" {
				t.Errorf("claims[\"iss\"] = %v, want %v", claims["iss"], "switchman")
			}
		})
	}
}

func TestProcessToken(t *testing.T) {
	tests := []struct {
		name    string
		sub     string
		iss     string
		iat     int64
		exp     int64
		secret  [32]byte
		wantErr bool // TODO: expect particular error
	}{
		{
			name:   "normal",
			sub:    "12345678",
			iss:    "switchman",
			iat:    time.Now().Unix(),
			exp:    time.Now().Add(time.Hour).Unix(),
			secret: testSecret,
		}, {
			name:    "bad_secret",
			sub:     "12345678",
			iss:     "switchman",
			iat:     time.Now().Unix(),
			exp:     time.Now().Add(time.Hour).Unix(),
			secret:  badSecret,
			wantErr: true,
		}, {
			name:    "bad_iss",
			sub:     "12345678",
			iss:     "notswitchman",
			iat:     time.Now().Unix(),
			exp:     time.Now().Add(time.Hour).Unix(),
			secret:  testSecret,
			wantErr: true,
		}, {
			name:    "expired",
			sub:     "12345678",
			iss:     "switchman",
			iat:     time.Now().Add(time.Hour * -2).Unix(),
			exp:     time.Now().Add(time.Hour * -1).Unix(),
			secret:  testSecret,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub": tt.sub,
				"iss": tt.iss,
				"exp": tt.exp,
				"iat": tt.iat,
			})
			tokenStr, err := token.SignedString(tt.secret[:])
			if err != nil {
				t.Errorf("token.SignedString: %v", err)
			}
			auth := makeTestAuth()
			gotid, err := auth.ProcessToken(tokenStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if gotid != tt.sub {
				t.Errorf("ProcessToken() = %v, want %v", gotid, tt.sub)
			}
		})
	}
}

func makeTestAuth() *AuthManager {
	settings := settings.VirtualStore{
		Data: map[string]interface{}{
			settingsPath: &authSettings{
				JwtSecret: testSecret[:],
			},
			usersPath: &map[string]*user{},
		},
	}
	auth, err := NewManager(&settings)
	if err != nil {
		panic(err)
	}
	return auth
}
