package auth

import (
	"fmt"

	"github.com/arrowinaknee/switchman/pkg/settings"
)

type AuthManager struct {
	store settings.Store
	Users *UserManager
}

func NewManager(store settings.Store) (*AuthManager, error) {
	users, err := NewUserManager(store)
	if err != nil {
		return nil, fmt.Errorf("auth: init manager: %w", err)
	}
	return &AuthManager{
		store: store,
		Users: users,
	}, nil
}
