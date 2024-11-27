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
			name:    "empty_password",
			creds:   userCreds{"login", ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &stubUserStore{data}
			users := &UserManager{
				store: store,
			}
			_ = users.load()

			u, err := users.Create(tt.creds.login, tt.creds.password)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("userManager.Create() error: %v", err)
				}
				if u != nil {
					t.Errorf("userManager.Create(): want nil user on error, got %v", u)
					u = nil
				}
			}

			if u != nil {
				if len(u.id) == 0 {
					t.Errorf("empty User.id")
				}
				if u.Login != tt.creds.login {
					t.Errorf("want login='%s', got '%s'", tt.creds.login, u.Login)
				}
				// TODO: password
				if len(u.Salt) == 0 {
					t.Errorf("empty User.Salt")
				}
			}

			wantUsers := cloneUsers(data.Users)
			if u != nil {
				wantUsers[u.id] = u
			}
			gotUsers := store.data.Users
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

func makeUsers(creds []userCreds) (users map[string]*User) {
	users = make(map[string]*User, len(creds))
	for _, c := range creds {
		var id string
		for {
			id = randomHexString()
			if _, ok := users[id]; !ok {
				break
			}
		}
		salt := randomHexString()

		users[id] = &User{
			id:        id,
			Login:     c.login,
			Password:  c.password,
			Salt:      salt,
			IsEnabled: true,
		}
	}
	return
}
func cloneUsers(in map[string]*User) (out map[string]*User) {
	out = make(map[string]*User, len(in))
	for i, u := range in {
		out[i] = new(User)
		*out[i] = *u
	}
	return
}
func checkUsers(want map[string]*User, got map[string]*User) (err error) {
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
