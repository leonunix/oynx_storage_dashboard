package auth

import (
	"path/filepath"
	"testing"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/store"
)

func TestDBUserStoreInitializeAndAuthenticate(t *testing.T) {
	db, err := store.OpenSQLite(filepath.Join(t.TempDir(), "dashboard.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	users, err := NewDBUserStore(db, "admin", "secret123", "admin")
	if err != nil {
		t.Fatalf("new user store: %v", err)
	}

	initialized, err := users.IsInitialized()
	if err != nil {
		t.Fatalf("check initialized: %v", err)
	}
	if initialized {
		t.Fatal("expected store to require initial setup")
	}

	_, err = users.Initialize(domain.SetupInitializeRequest{
		Username:    "admin",
		DisplayName: "Onyx Administrator",
		Password:    "secret123",
	})
	if err != nil {
		t.Fatalf("initialize store: %v", err)
	}

	user, err := users.Authenticate("admin", "secret123")
	if err != nil {
		t.Fatalf("authenticate initialized admin user: %v", err)
	}
	if user.Role != "admin" {
		t.Fatalf("expected admin role, got %q", user.Role)
	}
	if len(user.Permissions) == 0 {
		t.Fatal("expected permissions for bootstrap user")
	}
}

func TestDBUserStoreCRUDFlow(t *testing.T) {
	db, err := store.OpenSQLite(filepath.Join(t.TempDir(), "dashboard.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	users, err := NewDBUserStore(db, "admin", "secret123", "admin")
	if err != nil {
		t.Fatalf("new user store: %v", err)
	}

	_, err = users.Initialize(domain.SetupInitializeRequest{
		Username:    "admin",
		DisplayName: "Onyx Administrator",
		Password:    "secret123",
	})
	if err != nil {
		t.Fatalf("initialize store: %v", err)
	}

	created, err := users.Create(domain.UserCreateRequest{
		Username:    "alice",
		DisplayName: "Alice",
		Role:        "operator",
		Password:    "password1",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if created.Role != "operator" {
		t.Fatalf("expected operator role, got %q", created.Role)
	}

	updatedName := "Alice Ops"
	disabled := true
	updated, err := users.Update("alice", domain.UserUpdateRequest{
		DisplayName: &updatedName,
		Disabled:    &disabled,
	})
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if updated.DisplayName != updatedName {
		t.Fatalf("expected updated display name, got %q", updated.DisplayName)
	}
	if !updated.Disabled {
		t.Fatal("expected user to be disabled")
	}

	if err := users.ResetPassword("alice", "password2"); err != nil {
		t.Fatalf("reset password: %v", err)
	}

	if _, err := users.Authenticate("alice", "password2"); err == nil {
		t.Fatal("expected disabled user authentication to fail")
	}

	allUsers, err := users.List()
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(allUsers) != 2 {
		t.Fatalf("expected 2 users, got %d", len(allUsers))
	}
}
