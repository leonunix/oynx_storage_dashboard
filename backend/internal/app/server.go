package app

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/api"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/auth"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/config"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"
	appmw "github.com/leonunix/onyx_storage/dashboard/backend/internal/middleware"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/services"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/store"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/system"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/ui"
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
	onyxService := services.NewOnyxService(cfg.Onyx, runner)
	metricsHistory, err := services.NewMetricsHistoryService(cfg.Metrics, onyxService)
	if err != nil {
		return nil, fmt.Errorf("initialize metrics history service: %w", err)
	}

	handlers := &api.Handlers{
		UserStore:      userStore,
		JWTManager:     jwtManager,
		OnyxService:    onyxService,
		MetricsHistory: metricsHistory,
		StorageService: services.NewStorageService(cfg.Operations, runner, cfg.Command.StorageOpTimeout),
		ConfigService:  services.NewConfigService(cfg.Onyx.ConfigPath, cfg.Onyx.SocketPath, runner),
		AuditService:   auditService,
		SetupStatus: domain.SetupStatus{
			Initialized:       false,
			SuggestedUsername: cfg.Auth.BootstrapUsername,
			SuggestedRole:     cfg.Auth.BootstrapRole,
		},
	}

	// Embed frontend dist/ as SPA — sub to strip the "dist" prefix
	var frontendFS fs.FS
	if sub, err := fs.Sub(ui.DistFS, "dist"); err == nil {
		frontendFS = sub
	}

	router := api.NewRouter(api.RouterDependencies{
		AllowedOrigins: cfg.Server.AllowedOrigins,
		Handlers:       handlers,
		AuthMiddleware: appmw.Authenticator(jwtManager),
		FrontendFS:     frontendFS,
	})

	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}
	server.RegisterOnShutdown(func() {
		_ = metricsHistory.Stop()
	})
	metricsHistory.Start()

	return server, nil
}
