package users

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type UserManager struct {
	store store
	data  usersData
	mut   sync.Mutex
}

type usersData struct {
	JwtKey string           `yaml:"jwt_key"`
	Users  map[string]*User `yaml:"users"`
}

type User struct {
	id        string
	Login     string `yaml:"login"`
	Password  string `yaml:"password"`
	Salt      string `yaml:"salt"`
	IsEnabled bool   `yaml:"is_enabled"`
}

func InitManger(configPath string) (*UserManager, error) {
	m := &UserManager{
		store: getFileStore(configPath),
	}
	err := m.load()
	if err != nil {
		return nil, fmt.Errorf("init user manager: %v", err)
	}

	return m, nil
}

func (m *UserManager) load() error {
	d, err := m.store.loadData()
	if err != nil {
		return fmt.Errorf("load users config: %v", err)
	}
	m.data = *d

	for id, u := range m.data.Users {
		u.id = id
	}

	return nil
}
func (m *UserManager) save() error {
	if err := m.store.saveData(&m.data); err != nil {
		return fmt.Errorf("save users config: %v", err)
	}
	return nil
}

func (m *UserManager) Reload() error {
	m.mut.Lock()
	defer m.mut.Unlock()

	return m.load()
}

func (m *UserManager) Create(login string, password string) (*User, error) {
	m.mut.Lock()
	defer m.mut.Unlock()

	// TODO: custom errors for input validation
	if len(login) == 0 {
		return nil, fmt.Errorf("create user: empty login")
	}
	if len(password) == 0 {
		return nil, fmt.Errorf("create user: empty password")
	}

	var id string
	for {
		id = randomHexString()
		// make sure id does not repeat
		if _, ok := m.data.Users[id]; !ok {
			break
		}
	}

	salt := randomHexString()

	// FIXME: hashing
	pwHash := password

	m.data.Users[id] = &User{
		id:        id,
		Login:     login,
		Password:  pwHash,
		Salt:      salt,
		IsEnabled: true,
	}
	if err := m.save(); err != nil {
		return nil, fmt.Errorf("create user: %v", err)
	}

	return m.data.Users[id], nil
}

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomHexString() string {
	buf := make([]byte, 16)
	random.Read(buf) // random.Read() never returns errors

	return hex.EncodeToString(buf)
}
