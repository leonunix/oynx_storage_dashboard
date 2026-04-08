package auth

import (
	"errors"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"
)

type BootstrapUserStore struct {
	username string
	password string
	role     string
}

func NewBootstrapUserStore(username, password, role string) *BootstrapUserStore {
	return &BootstrapUserStore{
		username: username,
		password: password,
		role:     role,
	}
}

func (s *BootstrapUserStore) Authenticate(username, password string) (domain.User, error) {
	if username != s.username || password != s.password {
		return domain.User{}, errors.New("invalid credentials")
	}
	return domain.User{
		Username:    s.username,
		DisplayName: "Onyx Administrator",
		Role:        s.role,
		Permissions: PermissionsForRole(s.role),
	}, nil
}
