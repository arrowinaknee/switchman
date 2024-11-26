package users

import (
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
	JwtKey string          `yaml:"jwt_key"`
	Users  map[string]User `yaml:"users"`
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
		return nil, err
	}

	return m, nil
}

func (m *UserManager) load() error {
	d, err := m.store.loadData()
	if err != nil {
		return err
	}
	m.data = *d

	for id, u := range m.data.Users {
		u.id = id
	}

	return nil
}
func (m *UserManager) save() error {
	err := m.store.saveData(&m.data)
	return err
}

func (m *UserManager) Reload() error {
	m.mut.Lock()
	defer m.mut.Unlock()

	return m.load()
}

func (m *UserManager) Add(login string, password string) *User {
	m.mut.Lock()
	defer m.mut.Unlock()

	// TODO: validate credentials

}

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomHexString() string {
	buf := make([]byte, 16)
	if _, err := random.Read(); err != nil {

	}
}
