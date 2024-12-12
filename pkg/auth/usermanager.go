package auth

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/arrowinaknee/switchman/pkg/settings"
)

const usersPath = "users"

const minPasswordLen = 8

var (
	ErrLoginExists   = fmt.Errorf("auth: login already exists")
	ErrLoginEmpty    = fmt.Errorf("auth: login is empty")
	ErrLoginNotFound = fmt.Errorf("auth: user login not found")

	ErrPasswordTooShort = fmt.Errorf("auth: password too short")

	ErrUserNotFound = fmt.Errorf("auth: user not found")
	ErrUserDisabled = fmt.Errorf("auth: user is disabled")
)

type UserManager struct {
	store settings.Store
	users map[string]*user
	mut   sync.RWMutex
}

type user struct {
	Id        string `yaml:"-"`
	Login     string `yaml:"login"`
	Password  string `yaml:"password"`
	IsEnabled bool   `yaml:"is_enabled"`
}

func NewUserManager(store settings.Store) (*UserManager, error) {
	m := &UserManager{
		store: store,
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

func (m *UserManager) Create(login, password string) (string, error) {
	m.mut.Lock()
	defer m.mut.Unlock()

	if err := m.ValidateLogin(login); err != nil {
		return "", err
	}
	if err := m.ValidatePassword(password); err != nil {
		return "", err
	}

	var id string
	for {
		id = randomHexString()
		if _, ok := m.users[id]; !ok {
			break
		}
	}

	u := &user{
		Id:        id,
		Login:     login,
		Password:  createEncodedPassword(password),
		IsEnabled: true,
	}
	m.users[id] = u

	if err := m.saveData(); err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}

	return id, nil
}

func (m *UserManager) Delete(id string) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	if _, ok := m.users[id]; !ok {
		return ErrUserNotFound
	}

	delete(m.users, id)

	if err := m.saveData(); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}

func (m *UserManager) GetIdByLogin(login string) (string, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()

	u := m.findByLogin(login)
	if u == nil {
		return "", ErrLoginNotFound
	}
	return u.Id, nil
}

func (m *UserManager) GetUserLogin(id string) (string, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()

	u, ok := m.users[id]
	if !ok {
		return "", ErrUserNotFound
	}
	return u.Login, nil
}

func (m *UserManager) GetUserEnabled(id string) (bool, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()

	u, ok := m.users[id]
	if !ok {
		return false, ErrUserNotFound
	}
	return u.IsEnabled, nil
}

func (m *UserManager) SetUserPassword(id, password string) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	u, ok := m.users[id]
	if !ok {
		return ErrUserNotFound
	}

	if err := m.ValidatePassword(password); err != nil {
		return err
	}

	u.Password = createEncodedPassword(password)

	if err := m.saveData(); err != nil {
		return fmt.Errorf("set user password: %w", err)
	}
	return nil
}

func (m *UserManager) SetUserLogin(id, login string) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	u, ok := m.users[id]
	if !ok {
		return ErrUserNotFound
	}

	if err := m.ValidateLogin(login); err != nil {
		return err
	}

	u.Login = login

	if err := m.saveData(); err != nil {
		return fmt.Errorf("set user login: %w", err)
	}
	return nil
}

func (m *UserManager) SetUserEnabled(id string, enabled bool) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	u, ok := m.users[id]
	if !ok {
		return ErrUserNotFound
	}

	u.IsEnabled = enabled

	if err := m.saveData(); err != nil {
		return fmt.Errorf("set user enabled: %w", err)
	}
	return nil
}

func (m *UserManager) TrySignIn(login, password string) (string, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()

	u := m.findByLogin(login)
	if u == nil {
		return "", ErrLoginNotFound
	}

	if !u.IsEnabled {
		return "", ErrUserDisabled
	}

	if err := checkPassword(u.Password, password); err != nil {
		return "", err
	}

	return u.Id, nil
}

func (m *UserManager) ValidateLogin(login string) error {
	// TODO: do not allow spaces and special chars
	if login == "" {
		return ErrLoginEmpty
	}
	if lu := m.findByLogin(login); lu != nil {
		return ErrLoginExists
	}
	return nil
}

func (m *UserManager) ValidatePassword(password string) error {
	// TODO: more rules
	if len(password) < minPasswordLen {
		return ErrPasswordTooShort
	}
	return nil
}

func (m *UserManager) loadData() error {
	var users = map[string]*user{}
	err := m.store.Load(usersPath, &users)
	if err != nil {
		return fmt.Errorf("load users list: %w", err)
	}

	// TODO: data validation since users can be edited manually
	for id, u := range users {
		u.Id = id
	}

	m.users = users

	return nil
}
func (m *UserManager) saveData() error {
	if err := m.store.Save(usersPath, &m.users); err != nil {
		return fmt.Errorf("save users list: %w", err)
	}
	return nil
}

func (m *UserManager) findByLogin(login string) *user {
	for _, u := range m.users {
		if u.Login == login {
			return u
		}
	}

	return nil
}

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomHexString() string {
	buf := make([]byte, 16)
	random.Read(buf) // random.Read() never returns errors

	return hex.EncodeToString(buf)
}
