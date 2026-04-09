package api

import (
	"encoding/json"
	"fmt"
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
	ConfigService  *services.ConfigService
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

func (h *Handlers) MetricsSummary(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.OnyxService.MetricsJSON()
	if err != nil {
		// Fallback to overview which includes some metrics
		overview, oErr := h.OnyxService.Overview(r.Context())
		if oErr != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, overview)
		return
	}
	writeJSON(w, http.StatusOK, metrics)
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

// ── RAID management ─────────────────────────────────────────────────

func (h *Handlers) ListRaidArrays(w http.ResponseWriter, r *http.Request) {
	layout, err := h.StorageService.Layout(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": layout.RaidArrays})
}

func (h *Handlers) RaidDetail(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	arr, err := h.StorageService.RaidDetail(r.Context(), name)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, arr)
}

func (h *Handlers) CreateRaidArray(w http.ResponseWriter, r *http.Request) {
	var req domain.RaidCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	user, _ := appmw.UserFromContext(r.Context())

	if err := h.StorageService.RaidCreate(r.Context(), req); err != nil {
		status := http.StatusBadGateway
		if err == services.ErrMutationsDisabled {
			status = http.StatusForbidden
		}
		h.AuditService.Record(user.Username, "raid.create", req.Name, "error", err.Error())
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "raid.create", req.Name, "success", "RAID array created: "+req.Level)
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *Handlers) StopRaidArray(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	user, _ := appmw.UserFromContext(r.Context())

	if err := h.StorageService.RaidStop(r.Context(), domain.RaidStopRequest{Name: name}); err != nil {
		status := http.StatusBadGateway
		if err == services.ErrMutationsDisabled {
			status = http.StatusForbidden
		}
		h.AuditService.Record(user.Username, "raid.stop", name, "error", err.Error())
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "raid.stop", name, "success", "RAID array stopped")
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

// ── LVM PV management ──────────────────────────────────────────────

func (h *Handlers) CreatePV(w http.ResponseWriter, r *http.Request) {
	var req domain.PVCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	user, _ := appmw.UserFromContext(r.Context())

	if err := h.StorageService.PVCreate(r.Context(), req); err != nil {
		status := http.StatusBadGateway
		if err == services.ErrMutationsDisabled {
			status = http.StatusForbidden
		}
		h.AuditService.Record(user.Username, "pv.create", req.Device, "error", err.Error())
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "pv.create", req.Device, "success", "physical volume created")
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *Handlers) RemovePV(w http.ResponseWriter, r *http.Request) {
	var req domain.PVRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	user, _ := appmw.UserFromContext(r.Context())

	if err := h.StorageService.PVRemove(r.Context(), req); err != nil {
		status := http.StatusBadGateway
		if err == services.ErrMutationsDisabled {
			status = http.StatusForbidden
		}
		h.AuditService.Record(user.Username, "pv.remove", req.Device, "error", err.Error())
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "pv.remove", req.Device, "success", "physical volume removed")
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// ── LVM VG management ──────────────────────────────────────────────

func (h *Handlers) CreateVG(w http.ResponseWriter, r *http.Request) {
	var req domain.VGCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	user, _ := appmw.UserFromContext(r.Context())

	if err := h.StorageService.VGCreate(r.Context(), req); err != nil {
		status := http.StatusBadGateway
		if err == services.ErrMutationsDisabled {
			status = http.StatusForbidden
		}
		h.AuditService.Record(user.Username, "vg.create", req.Name, "error", err.Error())
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "vg.create", req.Name, "success", "volume group created")
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *Handlers) RemoveVG(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	user, _ := appmw.UserFromContext(r.Context())

	if err := h.StorageService.VGRemove(r.Context(), name, false); err != nil {
		status := http.StatusBadGateway
		if err == services.ErrMutationsDisabled {
			status = http.StatusForbidden
		}
		h.AuditService.Record(user.Username, "vg.remove", name, "error", err.Error())
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "vg.remove", name, "success", "volume group removed")
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// ── LVM LV management ──────────────────────────────────────────────

func (h *Handlers) CreateLV(w http.ResponseWriter, r *http.Request) {
	var req domain.LVCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	user, _ := appmw.UserFromContext(r.Context())

	if err := h.StorageService.LVCreate(r.Context(), req); err != nil {
		status := http.StatusBadGateway
		if err == services.ErrMutationsDisabled {
			status = http.StatusForbidden
		}
		h.AuditService.Record(user.Username, "lv.create", req.VGName+"/"+req.Name, "error", err.Error())
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "lv.create", req.VGName+"/"+req.Name, "success", "logical volume created")
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *Handlers) RemoveLV(w http.ResponseWriter, r *http.Request) {
	var req domain.LVRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	user, _ := appmw.UserFromContext(r.Context())

	if err := h.StorageService.LVRemove(r.Context(), req); err != nil {
		status := http.StatusBadGateway
		if err == services.ErrMutationsDisabled {
			status = http.StatusForbidden
		}
		h.AuditService.Record(user.Username, "lv.remove", req.VGName+"/"+req.Name, "error", err.Error())
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "lv.remove", req.VGName+"/"+req.Name, "success", "logical volume removed")
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (h *Handlers) ResizeLV(w http.ResponseWriter, r *http.Request) {
	var req domain.LVResizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	user, _ := appmw.UserFromContext(r.Context())

	if err := h.StorageService.LVResize(r.Context(), req); err != nil {
		status := http.StatusBadGateway
		if err == services.ErrMutationsDisabled {
			status = http.StatusForbidden
		}
		h.AuditService.Record(user.Username, "lv.resize", req.VGName+"/"+req.Name, "error", err.Error())
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "lv.resize", req.VGName+"/"+req.Name, "success", "logical volume resized to "+req.Size)
	writeJSON(w, http.StatusOK, map[string]string{"status": "resized"})
}

// ── Provision execution ────────────────────────────────────────────

func (h *Handlers) ExecuteProvision(w http.ResponseWriter, r *http.Request) {
	var req domain.ProvisionExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	user, _ := appmw.UserFromContext(r.Context())

	result, err := h.StorageService.ExecuteProvision(r.Context(), req)
	if err != nil {
		status := http.StatusBadGateway
		if err == services.ErrMutationsDisabled {
			status = http.StatusForbidden
		}
		h.AuditService.Record(user.Username, "storage.provision.execute", "", "error", err.Error())
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	resultStr := "success"
	if !result.Success {
		resultStr = "partial_failure"
	}
	h.AuditService.Record(user.Username, "storage.provision.execute", "", resultStr,
		fmt.Sprintf("executed %d commands, success=%v", len(result.Results), result.Success))
	writeJSON(w, http.StatusOK, result)
}

// ── Engine Configuration ────────────────────────────────────────────

func (h *Handlers) GetConfig(w http.ResponseWriter, _ *http.Request) {
	cfg, err := h.ConfigService.Read()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	mode, _ := h.ConfigService.Mode()

	writeJSON(w, http.StatusOK, map[string]any{
		"config": cfg,
		"mode":   mode,
	})
}

func (h *Handlers) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var cfg domain.EngineConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, _ := appmw.UserFromContext(r.Context())

	if err := h.ConfigService.Write(cfg); err != nil {
		h.AuditService.Record(user.Username, "config.update", "engine", "error", err.Error())
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.AuditService.Record(user.Username, "config.update", "engine", "success", "engine configuration updated")
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (h *Handlers) RestartService(w http.ResponseWriter, r *http.Request) {
	user, _ := appmw.UserFromContext(r.Context())
	h.AuditService.Record(user.Username, "config.restart", "engine", "success", "service restart requested")

	if err := h.ConfigService.RestartService(r.Context()); err != nil {
		h.AuditService.Record(user.Username, "config.restart", "engine", "error", err.Error())
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "restarting"})
}

func (h *Handlers) ReloadEngine(w http.ResponseWriter, r *http.Request) {
	user, _ := appmw.UserFromContext(r.Context())

	if err := h.ConfigService.Reload(); err != nil {
		h.AuditService.Record(user.Username, "config.reload", "engine", "error", err.Error())
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	mode, _ := h.ConfigService.Mode()
	h.AuditService.Record(user.Username, "config.reload", "engine", "success", "engine reloaded, mode: "+mode)
	writeJSON(w, http.StatusOK, domain.ConfigReloadResponse{
		Mode:    mode,
		Message: "engine reloaded successfully",
	})
}
