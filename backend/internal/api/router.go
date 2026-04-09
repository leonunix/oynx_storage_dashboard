package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	appmw "github.com/leonunix/onyx_storage/dashboard/backend/internal/middleware"
)

type RouterDependencies struct {
	AllowedOrigins []string
	Handlers       *Handlers
	AuthMiddleware func(http.Handler) http.Handler
}

func NewRouter(deps RouterDependencies) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   deps.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/setup/status", deps.Handlers.GetSetupStatus)
		r.Post("/setup/initialize", deps.Handlers.InitializeSetup)
		r.Post("/auth/login", deps.Handlers.Login)

		r.Group(func(protected chi.Router) {
			protected.Use(deps.AuthMiddleware)

			protected.Get("/auth/me", deps.Handlers.Me)

			protected.With(appmw.RequirePermission("overview:read")).Get("/dashboard/overview", deps.Handlers.Overview)
			protected.With(appmw.RequirePermission("metrics:read")).Get("/metrics/summary", deps.Handlers.Overview)
			protected.With(appmw.RequirePermission("volumes:read")).Get("/volumes", deps.Handlers.ListVolumes)
			protected.With(appmw.RequirePermission("volumes:write")).Post("/volumes", deps.Handlers.CreateVolume)
			protected.With(appmw.RequirePermission("volumes:write")).Delete("/volumes/{name}", deps.Handlers.DeleteVolume)
			protected.With(appmw.RequirePermission("storage:read")).Get("/storage/layout", deps.Handlers.StorageLayout)
			protected.With(appmw.RequirePermission("storage:write")).Post("/storage/workflows/provision/preview", deps.Handlers.ProvisionPreview)
			protected.With(appmw.RequirePermission("audit:read")).Get("/audit/events", deps.Handlers.ListAuditEvents)
			protected.With(appmw.RequirePermission("users:manage")).Get("/users", deps.Handlers.ListUsers)
			protected.With(appmw.RequirePermission("users:manage")).Post("/users", deps.Handlers.CreateUser)
			protected.With(appmw.RequirePermission("users:manage")).Patch("/users/{username}", deps.Handlers.UpdateUser)
			protected.With(appmw.RequirePermission("users:manage")).Post("/users/{username}/reset-password", deps.Handlers.ResetUserPassword)
			protected.With(appmw.RequirePermission("users:manage")).Get("/roles", deps.Handlers.ListRoles)
		})
	})

	return router
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
