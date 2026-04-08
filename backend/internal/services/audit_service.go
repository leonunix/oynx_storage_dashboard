package services

import (
	"sync"
	"time"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"
)

type AuditService struct {
	mu     sync.RWMutex
	events []domain.AuditEvent
}

func NewAuditService() *AuditService {
	return &AuditService{
		events: []domain.AuditEvent{
			{
				ID:          "bootstrap",
				At:          time.Now(),
				Actor:       "system",
				Action:      "dashboard.bootstrap",
				Resource:    "dashboard",
				Result:      "success",
				Description: "dashboard backend started with bootstrap administrator",
			},
		},
	}
}

func (s *AuditService) Record(actor, action, resource, result, description string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append([]domain.AuditEvent{{
		ID:          time.Now().UTC().Format(time.RFC3339Nano),
		At:          time.Now().UTC(),
		Actor:       actor,
		Action:      action,
		Resource:    resource,
		Result:      result,
		Description: description,
	}}, s.events...)

	if len(s.events) > 500 {
		s.events = s.events[:500]
	}
}

func (s *AuditService) List() []domain.AuditEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.AuditEvent, len(s.events))
	copy(out, s.events)
	return out
}
