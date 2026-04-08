package app

import (
	"net/http"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/api"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/auth"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/config"
	appmw "github.com/leonunix/onyx_storage/dashboard/backend/internal/middleware"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/services"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/system"
)

func NewServer(cfg config.Config) *http.Server {
	runner := system.NewRunner(cfg.Command.ExecTimeout)
	auditService := services.NewAuditService()
	userStore := auth.NewBootstrapUserStore(
		cfg.Auth.BootstrapUsername,
		cfg.Auth.BootstrapPassword,
		cfg.Auth.BootstrapRole,
	)
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.TokenTTL)

	handlers := &api.Handlers{
		UserStore:      userStore,
		JWTManager:     jwtManager,
		OnyxService:    services.NewOnyxService(cfg.Onyx, runner),
		StorageService: services.NewStorageService(cfg.Operations, runner),
		AuditService:   auditService,
	}

	router := api.NewRouter(api.RouterDependencies{
		AllowedOrigins: cfg.Server.AllowedOrigins,
		Handlers:       handlers,
		AuthMiddleware: appmw.Authenticator(jwtManager),
	})

	return &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}
}
