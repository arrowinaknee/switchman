package auth

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

var existingCreds = userCreds{"user1", "mypass"}                // must be present in default data
var secondCreds = userCreds{"tester", "sTR0n9er"}               // must be present in default data
var nonexistentCreds = userCreds{"nonexistent", "notapassword"} // must not be present in default data
var nonexistentId = "nonexistent"                               // id is not validated, can be anything

var data = usersData{
	Users: makeUsers([]userCreds{
		existingCreds,
		secondCreds,
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
		}, {
			name:    "duplicate_login",
			creds:   existingCreds,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// store never modifies initial data
			store := &stubUserStore{data}
			// TODO: more generic store, use normal constructor
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

			// if create returned an error, we will still check the user list
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

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "normal",
			id:      findId(data.Users, existingCreds.login),
			wantErr: false,
		}, {
			name:    "empty_id",
			id:      "",
			wantErr: true,
		}, {
			name:    "nonexistent_id",
			id:      nonexistentId,
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

			err := users.Delete(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("userManager.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			wantUsers := cloneUsers(data.Users)
			gotUsers := store.data.Users
			delete(wantUsers, tt.id)
			if err = checkUsers(wantUsers, gotUsers); err != nil {
				t.Errorf("user list is incorrect:\n%v", err)
			}
		})
	}
}

func TestGetIdByLogin(t *testing.T) {
	tests := []struct {
		name    string
		login   string
		wantId  string
		wantErr bool
	}{
		{
			name:    "normal",
			login:   existingCreds.login,
			wantId:  findId(data.Users, existingCreds.login),
			wantErr: false,
		}, {
			name:    "empty_login",
			login:   "",
			wantErr: true,
		}, {
			name:    "nonexistent_login",
			login:   nonexistentCreds.login,
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

			id, err := users.GetIdByLogin(tt.login)
			if (err != nil) != tt.wantErr {
				t.Errorf("userManager.GetIdByLogin() error = %v, wantErr %v", err, tt.wantErr)
			}
			if id != tt.wantId {
				t.Errorf("want id='%s', got '%s'", tt.wantId, id)
			}
		})
	}
}

func TestSetUserPassword(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		newPass string
		wantErr bool
	}{
		{
			name:    "normal",
			id:      findId(data.Users, existingCreds.login),
			newPass: "newPassword",
			wantErr: false,
		}, {
			name:    "invalid_password",
			id:      findId(data.Users, existingCreds.login),
			newPass: "short",
			wantErr: true,
		}, {
			name:    "empty_id",
			id:      "",
			wantErr: true,
		}, {
			name:    "nonexistent_id",
			id:      nonexistentId,
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

			err := users.SetUserPassword(tt.id, tt.newPass)
			if (err != nil) != tt.wantErr {
				t.Errorf("userManager.SetUserPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			login := store.data.Users[tt.id].Login
			lid, err := users.TrySignIn(login, tt.newPass)
			if err != nil {
				t.Errorf("userManager.TrySignIn() error = %v", err)
			}
			if lid != tt.id {
				t.Errorf("try sign in with new password: want id='%s', got '%s'", tt.id, lid)
			}
		})
	}
}

func TestSetUserLogin(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		newLogin string
		wantErr  bool
	}{
		{
			name:     "normal",
			id:       findId(data.Users, existingCreds.login),
			newLogin: "newLogin",
			wantErr:  false,
		}, {
			name:     "invalid_login",
			id:       findId(data.Users, existingCreds.login),
			newLogin: "",
			wantErr:  true,
		}, {
			name:     "duplicate_login",
			id:       findId(data.Users, existingCreds.login),
			newLogin: secondCreds.login,
			wantErr:  true,
		}, {
			name:    "empty_id",
			id:      "",
			wantErr: true,
		}, {
			name:    "nonexistent_id",
			id:      nonexistentId,
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

			err := users.SetUserLogin(tt.id, tt.newLogin)
			if (err != nil) != tt.wantErr {
				t.Errorf("userManager.SetUserLogin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			login, err := users.GetUserLogin(tt.id)
			if err != nil {
				t.Errorf("userManager.GetUserLogin() error = %v", err)
			}
			if login != tt.newLogin {
				t.Errorf("userManager.GetUserLogin(): want login='%s', got '%s'", tt.newLogin, login)
			}
		})
	}
}

func TestTrySignIn(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		password string
		wantId   string
		wantErr  bool
	}{
		{
			name:     "normal",
			login:    existingCreds.login,
			password: existingCreds.password,
			wantId:   findId(data.Users, existingCreds.login),
			wantErr:  false,
		}, {
			name:     "wrong_password",
			login:    existingCreds.login,
			password: "wrong",
			wantId:   "",
			wantErr:  true,
		}, {
			name:     "wrong_login",
			login:    "wrong",
			password: existingCreds.password,
			wantId:   "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &stubUserStore{data}
			users := &UserManager{
				store: store,
			}
			_ = users.loadData()

			id, err := users.TrySignIn(tt.login, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("userManager.TrySignIn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if id != tt.wantId {
				t.Errorf("userManager.TrySignIn(): want id='%s', got '%s'", tt.wantId, id)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "normal",
			password: "password",
			wantErr:  false,
		}, {
			name:     "short",
			password: "pass",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := &UserManager{}
			err := users.ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name    string
		login   string
		wantErr bool
	}{
		{
			name:    "normal",
			login:   nonexistentCreds.login,
			wantErr: false,
		}, {
			name:    "duplicate",
			login:   existingCreds.login,
			wantErr: true,
		}, {
			name:    "empty",
			login:   "",
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
			err := users.ValidateLogin(tt.login)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLogin() error = %v, wantErr %v", err, tt.wantErr)
				return
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

func findId(users map[string]*user, login string) string {
	for id, u := range users {
		if u.Login == login {
			return id
		}
	}
	return ""
}
