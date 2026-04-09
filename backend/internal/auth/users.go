package auth

import (
	"errors"
	"strings"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserStore interface {
	IsInitialized() (bool, error)
	Initialize(req domain.SetupInitializeRequest) (domain.User, error)
	Authenticate(username, password string) (domain.User, error)
	List() ([]domain.User, error)
	Create(req domain.UserCreateRequest) (domain.User, error)
	Update(username string, req domain.UserUpdateRequest) (domain.User, error)
	ResetPassword(username, password string) error
}

type UserRecord struct {
	Username     string `gorm:"primaryKey;size:128"`
	DisplayName  string `gorm:"size:255;not null"`
	Role         string `gorm:"size:64;not null"`
	PasswordHash string `gorm:"size:255;not null"`
	Disabled     bool   `gorm:"not null;default:false"`
}

type DBUserStore struct {
	db *gorm.DB
}

func NewDBUserStore(db *gorm.DB, bootstrapUsername, bootstrapPassword, bootstrapRole string) (*DBUserStore, error) {
	store := &DBUserStore{db: db}
	if err := db.AutoMigrate(&UserRecord{}); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *DBUserStore) IsInitialized() (bool, error) {
	var count int64
	if err := s.db.Model(&UserRecord{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *DBUserStore) Initialize(req domain.SetupInitializeRequest) (domain.User, error) {
	initialized, err := s.IsInitialized()
	if err != nil {
		return domain.User{}, err
	}
	if initialized {
		return domain.User{}, errors.New("dashboard already initialized")
	}

	username := strings.TrimSpace(req.Username)
	if username == "" {
		return domain.User{}, errors.New("username is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return domain.User{}, errors.New("password is required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}

	record := UserRecord{
		Username:     username,
		DisplayName:  strings.TrimSpace(req.DisplayName),
		Role:         "admin",
		PasswordHash: string(hash),
		Disabled:     false,
	}
	if record.DisplayName == "" {
		record.DisplayName = username
	}

	if err := s.db.Create(&record).Error; err != nil {
		return domain.User{}, err
	}
	return toDomainUser(record), nil
}

func (s *DBUserStore) Authenticate(username, password string) (domain.User, error) {
	initialized, err := s.IsInitialized()
	if err != nil {
		return domain.User{}, err
	}
	if !initialized {
		return domain.User{}, errors.New("dashboard setup required")
	}

	var record UserRecord
	if err := s.db.Where("username = ?", strings.TrimSpace(username)).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, errors.New("invalid credentials")
		}
		return domain.User{}, err
	}

	if record.Disabled {
		return domain.User{}, errors.New("user disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(record.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, errors.New("invalid credentials")
	}

	return domain.User{
		Username:    record.Username,
		DisplayName: record.DisplayName,
		Role:        record.Role,
		Permissions: PermissionsForRole(record.Role),
		Disabled:    record.Disabled,
	}, nil
}

func (s *DBUserStore) List() ([]domain.User, error) {
	var records []UserRecord
	if err := s.db.Order("username asc").Find(&records).Error; err != nil {
		return nil, err
	}

	users := make([]domain.User, 0, len(records))
	for _, record := range records {
		users = append(users, toDomainUser(record))
	}
	return users, nil
}

func (s *DBUserStore) Create(req domain.UserCreateRequest) (domain.User, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return domain.User{}, errors.New("username is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return domain.User{}, errors.New("password is required")
	}

	role := normalizeRole(req.Role)
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}

	record := UserRecord{
		Username:     username,
		DisplayName:  strings.TrimSpace(req.DisplayName),
		Role:         role,
		PasswordHash: string(hash),
		Disabled:     false,
	}
	if record.DisplayName == "" {
		record.DisplayName = username
	}

	if err := s.db.Create(&record).Error; err != nil {
		return domain.User{}, err
	}
	return toDomainUser(record), nil
}

func (s *DBUserStore) Update(username string, req domain.UserUpdateRequest) (domain.User, error) {
	var record UserRecord
	if err := s.db.Where("username = ?", strings.TrimSpace(username)).First(&record).Error; err != nil {
		return domain.User{}, err
	}

	if req.DisplayName != nil {
		value := strings.TrimSpace(*req.DisplayName)
		if value != "" {
			record.DisplayName = value
		}
	}
	if req.Role != nil {
		record.Role = normalizeRole(*req.Role)
	}
	if req.Disabled != nil {
		record.Disabled = *req.Disabled
	}

	if err := s.db.Save(&record).Error; err != nil {
		return domain.User{}, err
	}
	return toDomainUser(record), nil
}

func (s *DBUserStore) ResetPassword(username, password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("password is required")
	}

	var record UserRecord
	if err := s.db.Where("username = ?", strings.TrimSpace(username)).First(&record).Error; err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	record.PasswordHash = string(hash)
	return s.db.Save(&record).Error
}

func normalizeRole(role string) string {
	role = strings.TrimSpace(role)
	if _, ok := rolePermissions[role]; ok {
		return role
	}
	return "viewer"
}

func toDomainUser(record UserRecord) domain.User {
	return domain.User{
		Username:    record.Username,
		DisplayName: record.DisplayName,
		Role:        record.Role,
		Permissions: PermissionsForRole(record.Role),
		Disabled:    record.Disabled,
	}
}
