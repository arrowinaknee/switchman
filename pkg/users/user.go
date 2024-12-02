package users

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const minPasswordLen = 8

var ErrorPasswordTooShort = fmt.Errorf("auth: password too short")

type UserManager struct {
	store store
	data  usersData
	mut   sync.RWMutex
}

type usersData struct {
	JwtKey string           `yaml:"jwt_key"`
	Users  map[string]*User `yaml:"users"`
}

// TODO: do not export
type User struct {
	Id        string `yaml:"-"`
	Login     string `yaml:"login"`
	Password  string `yaml:"password"`
	IsEnabled bool   `yaml:"is_enabled"`
}

func InitManger(configPath string) (*UserManager, error) {
	m := &UserManager{
		store: getFileStore(configPath),
	}
	err := m.loadData()
	if err != nil {
		return nil, fmt.Errorf("init user manager: %v", err)
	}

	return m, nil
}

func (m *UserManager) Reload() error {
	m.mut.Lock()
	defer m.mut.Unlock()

	return m.loadData()
}

func (m *UserManager) GetById(id string) *User {
	m.mut.RLock()
	defer m.mut.RUnlock()

	return m.data.Users[id]
}

func (m *UserManager) findByLogin(login string) *User {
	for _, u := range m.data.Users {
		if u.Login == login {
			return u
		}
	}

	return nil
}

// Save stores user data and assigns new id to user if id is empty
func (m *UserManager) Save(u *User) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	isNew := false
	if u.Id == "" {
		isNew = true
		for {
			id := randomHexString()
			if _, ok := m.data.Users[id]; !ok {
				u.Id = id
				break
			}
		}
	}

	if u.Login == "" {
		return fmt.Errorf("save user: empty login")
	}
	if lu := m.findByLogin(u.Login); lu != nil && (isNew || lu.Id != u.Id) {
		return fmt.Errorf("save user: login already exists")
	}
	if err := verifyEncodedPassword(u.Password); err != nil {
		return fmt.Errorf("save user: %w", err)
	}

	m.data.Users[u.Id] = u

	err := m.saveData()
	if err != nil {
		return fmt.Errorf("save user: %w", err)
	}

	return nil
}

// NewEncodedPassword validates password and returns encoded password
func (m *UserManager) NewEncodedPassword(password string) (string, error) {
	if err := m.validatePassword(password); err != nil {
		return "", err
	}
	return createEncodedPassword(password), nil
}

// ValidatePassword checks password and returns error reason for password invalidaty
func (m *UserManager) ValidatePassword(password string) error {
	return m.validatePassword(password)
}

// MatchPassword checks if password matches user encoded password
func (u *User) MatchPassword(password string) bool {
	err := checkPassword(u.Password, password)
	if err == nil {
		return true
	}
	// user has invadid encoded password
	if !errors.Is(err, ErrPasswordMismatch) {
		// TODO: write to log
		fmt.Println("Error: check password - %v", err)
	}
	return false
}

func (m *UserManager) loadData() error {
	d, err := m.store.loadData()
	if err != nil {
		return fmt.Errorf("load users config: %v", err)
	}
	m.data = *d

	for id, u := range m.data.Users {
		u.Id = id
	}

	return nil
}
func (m *UserManager) saveData() error {
	if err := m.store.saveData(&m.data); err != nil {
		return fmt.Errorf("save users config: %v", err)
	}
	return nil
}

// TODO: more rules and move to separate file
func (m *UserManager) validatePassword(password string) error {
	if len(password) < minPasswordLen {
		return ErrorPasswordTooShort
	}
	return nil
}

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomHexString() string {
	buf := make([]byte, 16)
	random.Read(buf) // random.Read() never returns errors

	return hex.EncodeToString(buf)
}
