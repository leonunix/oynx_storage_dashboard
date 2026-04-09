package app

import (
	"fmt"
	"net/http"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/api"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/auth"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/config"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"
	appmw "github.com/leonunix/onyx_storage/dashboard/backend/internal/middleware"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/services"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/store"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/system"
)

func NewServer(cfg config.Config) (*http.Server, error) {
	db, err := store.OpenSQLite(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("open dashboard database: %w", err)
	}

	runner := system.NewRunner(cfg.Command.ExecTimeout)
	auditService, err := services.NewAuditService(db)
	if err != nil {
		return nil, fmt.Errorf("initialize audit service: %w", err)
	}
	userStore, err := auth.NewDBUserStore(
		db,
		cfg.Auth.BootstrapUsername,
		cfg.Auth.BootstrapPassword,
		cfg.Auth.BootstrapRole,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize user store: %w", err)
	}
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.TokenTTL)

	handlers := &api.Handlers{
		UserStore:      userStore,
		JWTManager:     jwtManager,
		OnyxService:    services.NewOnyxService(cfg.Onyx, runner),
		StorageService: services.NewStorageService(cfg.Operations, runner),
		AuditService:   auditService,
		SetupStatus: domain.SetupStatus{
			Initialized:       false,
			SuggestedUsername: cfg.Auth.BootstrapUsername,
			SuggestedRole:     cfg.Auth.BootstrapRole,
		},
	}

	router := api.NewRouter(api.RouterDependencies{
		AllowedOrigins: cfg.Server.AllowedOrigins,
		Handlers:       handlers,
		AuthMiddleware: appmw.Authenticator(jwtManager),
	})

	return &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}, nil
}
