package services

import (
	"time"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"
	"gorm.io/gorm"
)

type AuditService struct {
	db *gorm.DB
}

type AuditEventRecord struct {
	ID          string    `gorm:"primaryKey;size:64"`
	At          time.Time `gorm:"index;not null"`
	Actor       string    `gorm:"size:128;index;not null"`
	Action      string    `gorm:"size:128;index;not null"`
	Resource    string    `gorm:"size:255;index;not null"`
	Result      string    `gorm:"size:32;index;not null"`
	Description string    `gorm:"type:text;not null"`
}

func NewAuditService(db *gorm.DB) (*AuditService, error) {
	if err := db.AutoMigrate(&AuditEventRecord{}); err != nil {
		return nil, err
	}

	service := &AuditService{db: db}
	var count int64
	if err := db.Model(&AuditEventRecord{}).Count(&count).Error; err == nil && count == 0 {
		service.Record(
			"system",
			"dashboard.bootstrap",
			"dashboard",
			"success",
			"dashboard backend started with bootstrap administrator",
		)
	}
	return service, nil
}

func (s *AuditService) Record(actor, action, resource, result, description string) {
	_ = s.db.Create(&AuditEventRecord{
		ID:          time.Now().UTC().Format(time.RFC3339Nano),
		At:          time.Now().UTC(),
		Actor:       actor,
		Action:      action,
		Resource:    resource,
		Result:      result,
		Description: description,
	}).Error
}

func (s *AuditService) List() ([]domain.AuditEvent, error) {
	var records []AuditEventRecord
	if err := s.db.Order("at desc").Limit(500).Find(&records).Error; err != nil {
		return nil, err
	}

	out := make([]domain.AuditEvent, 0, len(records))
	for _, record := range records {
		out = append(out, domain.AuditEvent{
			ID:          record.ID,
			At:          record.At,
			Actor:       record.Actor,
			Action:      record.Action,
			Resource:    record.Resource,
			Result:      record.Result,
			Description: record.Description,
		})
	}
	return out, nil
}
