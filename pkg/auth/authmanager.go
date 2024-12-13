package auth

import (
	"fmt"
	"sync"
	"time"

	"github.com/arrowinaknee/switchman/pkg/settings"
	"github.com/golang-jwt/jwt/v5"
)

const settingsPath = "config.auth"

type AuthManager struct {
	mut      sync.RWMutex
	store    settings.Store
	Users    *UserManager
	settings authSettings
}

type authSettings struct {
	JwtSecret []byte `yaml:"jwt_secret"`
}

func NewManager(store settings.Store) (*AuthManager, error) {
	users, err := NewUserManager(store)
	if err != nil {
		return nil, fmt.Errorf("auth: init manager: %w", err)
	}
	m := &AuthManager{
		store: store,
		Users: users,
	}
	// TODO: allow notFound, set defaults and generate secret
	if err := m.loadSettings(); err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	return m, nil
}

func (m *AuthManager) Reload() error {
	m.mut.Lock()
	defer m.mut.Unlock()

	if err := m.loadSettings(); err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	return nil
}

func (m *AuthManager) IssueToken(userId string) (string, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId,
		"iss": "switchman",
		"exp": time.Now().Add(time.Hour * 2).Unix(),
		"iat": time.Now().Unix(),
	})

	token, err := claims.SignedString(m.settings.JwtSecret)
	if err != nil {
		return "", fmt.Errorf("issue token: %w", err)
	}
	return token, nil
}

func (m *AuthManager) ProcessToken(tokenString string) (userId string, err error) {
	m.mut.RLock()
	defer m.mut.RUnlock()

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return m.settings.JwtSecret, nil
	})
	if err != nil {
		return "", fmt.Errorf("process token: %w", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("process token: invalid token")
	}
	userId = claims["sub"].(string)
	return userId, nil
}

func (m *AuthManager) loadSettings() error {
	err := m.store.Load(settingsPath, &m.settings)
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	return nil
}

func (m *AuthManager) saveSettings() error {
	if err := m.store.Save(settingsPath, &m.settings); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}
	return nil
}
