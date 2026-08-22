package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/codingguna/aio-panel/internal/db"
	"github.com/codingguna/aio-panel/internal/system"
)

type ServicesHandler struct {
	store *db.Store
}

func NewServicesHandler(store *db.Store) *ServicesHandler {
	return &ServicesHandler{store: store}
}

// ListServices handles GET /api/v1/services
func (h *ServicesHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	services, err := system.ListServices(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list services"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// InspectService handles GET /api/v1/services/{name}
func (h *ServicesHandler) InspectService(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, `{"error":"missing service name"}`, http.StatusBadRequest)
		return
	}

	svc, err := system.InspectService(r.Context(), name)
	if err != nil {
		http.Error(w, `{"error":"service not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(svc)
}

type ServiceActionRequest struct {
	Action string `json:"action"` // start, stop, restart, reload, enable, disable
}

type ServiceActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Service string `json:"service"`
	Action  string `json:"action"`
}

// ControlService handles POST /api/v1/services/{name}/action
func (h *ServicesHandler) ControlService(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, `{"error":"missing service name"}`, http.StatusBadRequest)
		return
	}

	var req ServiceActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	action := strings.ToLower(strings.TrimSpace(req.Action))
	err := system.ControlService(r.Context(), name, action)

	resultStatus := "SUCCESS"
	var errDetails string
	if err != nil {
		resultStatus = "FAILURE"
		errDetails = err.Error()
	}

	// Audit logging
	if h.store != nil {
		clientIP := r.RemoteAddr
		_ = h.store.LogAudit(r.Context(), "admin", strings.ToUpper(action)+"_SERVICE", name, resultStatus, errDetails, clientIP)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ServiceActionResponse{
			Success: false,
			Message: err.Error(),
			Service: name,
			Action:  action,
		})
		return
	}

	json.NewEncoder(w).Encode(ServiceActionResponse{
		Success: true,
		Message: "Action " + action + " executed successfully on " + name,
		Service: name,
		Action:  action,
	})
}

// GetServiceLogs handles GET /api/v1/services/{name}/logs
func (h *ServicesHandler) GetServiceLogs(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, `{"error":"missing service name"}`, http.StatusBadRequest)
		return
	}

	lines := 50
	if lStr := r.URL.Query().Get("lines"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			lines = l
		}
	}

	logs, err := system.GetServiceLogs(r.Context(), name, lines)
	if err != nil {
		http.Error(w, `{"error":"failed to retrieve logs"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(logs))
}

type CreateServiceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ExecStart   string `json:"exec_start"`
	WorkDir     string `json:"work_dir"`
	User        string `json:"user"`
	Restart     string `json:"restart"`
	Enable      bool   `json:"enable"`
}

// CreateService handles POST /api/v1/services/create
func (h *ServicesHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.ExecStart == "" {
		http.Error(w, `{"error":"service name and exec_start command are required"}`, http.StatusBadRequest)
		return
	}

	err := system.CreateService(r.Context(), req.Name, req.Description, req.ExecStart, req.WorkDir, req.User, req.Restart, req.Enable)
	resultStatus := "SUCCESS"
	var errDetails string
	if err != nil {
		resultStatus = "FAILURE"
		errDetails = err.Error()
	}

	if h.store != nil {
		_ = h.store.LogAudit(r.Context(), "admin", "CREATE_SERVICE", req.Name, resultStatus, errDetails, r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message": "Systemd service created successfully: " + req.Name + ".service",
		"service": req.Name,
	})
}

// DeleteService handles DELETE /api/v1/services/{name}
func (h *ServicesHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, `{"error":"missing service name"}`, http.StatusBadRequest)
		return
	}

	err := system.DeleteService(r.Context(), name)
	resultStatus := "SUCCESS"
	var errDetails string
	if err != nil {
		resultStatus = "FAILURE"
		errDetails = err.Error()
	}

	if h.store != nil {
		_ = h.store.LogAudit(r.Context(), "admin", "DELETE_SERVICE", name, resultStatus, errDetails, r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message": "Service " + name + " stopped and deleted successfully",
	})
}
