package services

import (
	"path/filepath"
	"testing"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/store"
)

func TestAuditServicePersistsEvents(t *testing.T) {
	db, err := store.OpenSQLite(filepath.Join(t.TempDir(), "dashboard.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	service, err := NewAuditService(db)
	if err != nil {
		t.Fatalf("new audit service: %v", err)
	}

	service.Record("admin", "volume.create", "vol-a", "success", "created volume")

	events, err := service.List()
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(events))
	}
	if events[0].Actor != "admin" {
		t.Fatalf("expected latest actor admin, got %q", events[0].Actor)
	}
}
