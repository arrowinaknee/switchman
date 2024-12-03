package users

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

var data = usersData{
	Users: makeUsers([]userCreds{
		{"user1", "mypass"},
		{"tester", "sTR0n9er"},
		{"super", "abcdefg1"},
	}),
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name    string
		creds   userCreds
		wantErr bool
	}{
		{
			name:    "normal",
			creds:   userCreds{"login", "password"},
			wantErr: false,
		}, {
			name:    "empty_login",
			creds:   userCreds{"", "password"},
			wantErr: true,
		}, {
			name:    "short_password",
			creds:   userCreds{"login", "pass"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &stubUserStore{data}
			users := &UserManager{
				store: store,
			}
			_ = users.loadData()

			id, err := users.Create(tt.creds.login, tt.creds.password)
			userCreated := true
			if err != nil {
				if !tt.wantErr {
					t.Errorf("userManager.Create() error: %v", err)
				}
				userCreated = false
			}

			if userCreated {
				if len(id) == 0 {
					t.Errorf("empty user id")
				}
				login, err := users.GetUserLogin(id)
				if err != nil {
					t.Errorf("userManager.GetUserLogin() error: %v", err)
				}
				if login != tt.creds.login {
					t.Errorf("want login='%s', got '%s'", tt.creds.login, login)
				}
				credId, err := users.TrySignIn(tt.creds.login, tt.creds.password)
				if err != nil {
					t.Errorf("userManager.TrySignIn() error: %v", err)
				}
				if credId != id {
					t.Errorf("want id='%s', got '%s'", id, credId)
				}
			}

			wantUsers := cloneUsers(data.Users)
			gotUsers := store.data.Users
			if userCreated {
				wantUsers[id] = gotUsers[id]
			}
			if err = checkUsers(wantUsers, gotUsers); err != nil {
				t.Errorf("user list is incorrect:\n%v", err)
			}
		})
	}
}

type stubUserStore struct {
	data usersData
}

func (s *stubUserStore) loadData() (*usersData, error) {
	return &usersData{
		JwtKey: s.data.JwtKey,
		Users:  cloneUsers(s.data.Users),
	}, nil
}

func (s *stubUserStore) saveData(d *usersData) error {
	s.data = usersData{
		JwtKey: d.JwtKey,
		Users:  cloneUsers(d.Users),
	}
	return nil
}

type userCreds struct {
	login    string
	password string
}

func makeUsers(creds []userCreds) (users map[string]*user) {
	users = make(map[string]*user, len(creds))
	for _, c := range creds {
		var id string
		for {
			id = randomHexString()
			if _, ok := users[id]; !ok {
				break
			}
		}

		users[id] = &user{
			Id:        id,
			Login:     c.login,
			Password:  createEncodedPassword(c.password),
			IsEnabled: true,
		}
	}
	return
}
func cloneUsers(in map[string]*user) (out map[string]*user) {
	out = make(map[string]*user, len(in))
	for i, u := range in {
		out[i] = new(user)
		*out[i] = *u
	}
	return
}
func checkUsers(want map[string]*user, got map[string]*user) (err error) {
	// could use reflect.DeepEqual at top level, but per element comparison is more informative
	for id, wantU := range want {
		gotU, ok := got[id]
		if !ok {
			err = errors.Join(err, fmt.Errorf("missing user %+v", wantU))
			continue
		}
		if !reflect.DeepEqual(wantU, gotU) {
			err = errors.Join(err, fmt.Errorf("users do not match:\n  want %+v\n  got  %+v", wantU, gotU))
		}
	}
	for id, gotU := range got {
		if _, ok := want[id]; !ok {
			err = errors.Join(err, fmt.Errorf("unexpected user %+v", gotU))
		}
	}
	return
}
