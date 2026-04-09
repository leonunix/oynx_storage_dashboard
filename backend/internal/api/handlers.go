package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/auth"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"
	appmw "github.com/leonunix/onyx_storage/dashboard/backend/internal/middleware"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/services"
)

type Handlers struct {
	UserStore      auth.UserStore
	JWTManager     *auth.JWTManager
	OnyxService    *services.OnyxService
	StorageService *services.StorageService
	AuditService   *services.AuditService
	SetupStatus    domain.SetupStatus
}

func (h *Handlers) GetSetupStatus(w http.ResponseWriter, _ *http.Request) {
	initialized, err := h.UserStore.IsInitialized()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	status := h.SetupStatus
	status.Initialized = initialized
	writeJSON(w, http.StatusOK, status)
}

func (h *Handlers) InitializeSetup(w http.ResponseWriter, r *http.Request) {
	var req domain.SetupInitializeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, err := h.UserStore.Initialize(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "setup.initialize", "dashboard", "success", "dashboard initialized")
	writeJSON(w, http.StatusCreated, user)
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, err := h.UserStore.Authenticate(strings.TrimSpace(req.Username), req.Password)
	if err != nil {
		if err.Error() == "dashboard setup required" {
			writeJSON(w, http.StatusPreconditionRequired, map[string]string{
				"error": "dashboard setup required",
				"code":  "setup_required",
			})
			return
		}
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		h.AuditService.Record(req.Username, "auth.login", "dashboard", "denied", "invalid credentials")
		return
	}

	token, err := h.JWTManager.Issue(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue token"})
		return
	}

	h.AuditService.Record(user.Username, "auth.login", "dashboard", "success", "dashboard login")
	writeJSON(w, http.StatusOK, domain.LoginResponse{Token: token, User: user})
}

func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	user, _ := appmw.UserFromContext(r.Context())
	writeJSON(w, http.StatusOK, user)
}

func (h *Handlers) Overview(w http.ResponseWriter, r *http.Request) {
	overview, err := h.OnyxService.Overview(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, overview)
}

func (h *Handlers) ListVolumes(w http.ResponseWriter, r *http.Request) {
	volumes, err := h.OnyxService.ListVolumes(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": volumes})
}

func (h *Handlers) CreateVolume(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Compression == "" {
		req.Compression = "lz4"
	}

	user, _ := appmw.UserFromContext(r.Context())
	if err := h.OnyxService.CreateVolume(r.Context(), req); err != nil {
		h.AuditService.Record(user.Username, "volume.create", req.Name, "error", err.Error())
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "volume.create", req.Name, "success", "volume created")
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *Handlers) DeleteVolume(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	user, _ := appmw.UserFromContext(r.Context())

	if err := h.OnyxService.DeleteVolume(r.Context(), name); err != nil {
		h.AuditService.Record(user.Username, "volume.delete", name, "error", err.Error())
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "volume.delete", name, "success", "volume deleted")
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handlers) StorageLayout(w http.ResponseWriter, r *http.Request) {
	layout, err := h.StorageService.Layout(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, layout)
}

func (h *Handlers) ProvisionPreview(w http.ResponseWriter, r *http.Request) {
	var req domain.ProvisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	user, _ := appmw.UserFromContext(r.Context())
	h.AuditService.Record(user.Username, "storage.provision.preview", req.Name, "success", "generated provisioning plan")
	writeJSON(w, http.StatusOK, h.StorageService.PlanProvision(r.Context(), req))
}

func (h *Handlers) ListAuditEvents(w http.ResponseWriter, _ *http.Request) {
	items, err := h.AuditService.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handlers) ListUsers(w http.ResponseWriter, _ *http.Request) {
	items, err := h.UserStore.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req domain.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, _ := appmw.UserFromContext(r.Context())
	created, err := h.UserStore.Create(req)
	if err != nil {
		h.AuditService.Record(user.Username, "user.create", req.Username, "error", err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "user.create", created.Username, "success", "user created")
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	var req domain.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, _ := appmw.UserFromContext(r.Context())
	updated, err := h.UserStore.Update(username, req)
	if err != nil {
		h.AuditService.Record(user.Username, "user.update", username, "error", err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "user.update", username, "success", "user updated")
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handlers) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	var req domain.UserPasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, _ := appmw.UserFromContext(r.Context())
	if err := h.UserStore.ResetPassword(username, req.Password); err != nil {
		h.AuditService.Record(user.Username, "user.reset_password", username, "error", err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "user.reset_password", username, "success", "password reset")
	writeJSON(w, http.StatusOK, map[string]string{"status": "password-reset"})
}

func (h *Handlers) ListRoles(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": auth.RoleDefinitions()})
}
