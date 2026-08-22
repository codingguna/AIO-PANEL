package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/codingguna/aio-panel/internal/config"
	"github.com/codingguna/aio-panel/internal/db"
	"github.com/codingguna/aio-panel/internal/security"
)

type SecurityHandler struct {
	cfg   *config.Config
	store *db.Store
}

func NewSecurityHandler(cfg *config.Config, store *db.Store) *SecurityHandler {
	return &SecurityHandler{cfg: cfg, store: store}
}

// GetSSHConfig handles GET /api/v1/security/ssh
func (h *SecurityHandler) GetSSHConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := security.GetSSHConfig(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to read sshd config"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

type UpdateSSHRequest struct {
	Port                   int    `json:"port"`
	PermitRootLogin        string `json:"permit_root_login"`
	PasswordAuthentication bool   `json:"password_authentication"`
}

// UpdateSSHConfig handles POST /api/v1/security/ssh/config
func (h *SecurityHandler) UpdateSSHConfig(w http.ResponseWriter, r *http.Request) {
	var req UpdateSSHRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.Port <= 0 || req.Port > 65535 {
		http.Error(w, `{"error":"invalid port number"}`, http.StatusBadRequest)
		return
	}

	err := security.UpdateSSHConfig(r.Context(), req.Port, req.PermitRootLogin, req.PasswordAuthentication, h.cfg.Paths.BackupDir)
	resultStatus := "SUCCESS"
	var errDetails string
	if err != nil {
		resultStatus = "FAILURE"
		errDetails = err.Error()
	}

	if h.store != nil {
		_ = h.store.LogAudit(r.Context(), "admin", "UPDATE_SSH_CONFIG", "/etc/ssh/sshd_config", resultStatus, errDetails, r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "SSH configuration updated, validated, and reloaded successfully.",
	})
}

// GetAuthorizedKeys handles GET /api/v1/security/ssh/keys
func (h *SecurityHandler) GetAuthorizedKeys(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		user = "root"
	}

	keys, err := security.ListAuthorizedKeys(user)
	if err != nil {
		http.Error(w, `{"error":"failed to list ssh keys"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

type AddSSHKeyRequest struct {
	User   string `json:"user"`
	Key    string `json:"key"`
}

// AddAuthorizedKey handles POST /api/v1/security/ssh/keys
func (h *SecurityHandler) AddAuthorizedKey(w http.ResponseWriter, r *http.Request) {
	var req AddSSHKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	err := security.AddAuthorizedKey(req.User, req.Key)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "ADD_SSH_KEY", req.User, status, "", r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "SSH Key added successfully."})
}

// GetSSHSessions handles GET /api/v1/security/ssh/sessions
func (h *SecurityHandler) GetSSHSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := security.GetActiveSSHSessions(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to get ssh sessions"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// GetFirewallStatus handles GET /api/v1/security/firewall
func (h *SecurityHandler) GetFirewallStatus(w http.ResponseWriter, r *http.Request) {
	fw, err := security.GetFirewallStatus(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to get firewall status"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fw)
}

type AddFirewallRuleRequest struct {
	Port     string `json:"port"`
	Protocol string `json:"protocol"` // tcp, udp, any
	Action   string `json:"action"`   // ALLOW, DENY
	FromIP   string `json:"from_ip"`
	Comment  string `json:"comment"`
}

// AddFirewallRule handles POST /api/v1/security/firewall/rules
func (h *SecurityHandler) AddFirewallRule(w http.ResponseWriter, r *http.Request) {
	var req AddFirewallRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.Action == "" {
		req.Action = "ALLOW"
	}

	err := security.AddFirewallRule(r.Context(), req.Port, req.Protocol, req.Action, req.FromIP, req.Comment)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "ADD_FIREWALL_RULE", req.Port+"/"+req.Protocol, status, req.Comment, r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Firewall rule created successfully."})
}

// DeleteFirewallRule handles DELETE /api/v1/security/firewall/rules/{id}
func (h *SecurityHandler) DeleteFirewallRule(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid rule ID"}`, http.StatusBadRequest)
		return
	}

	err = security.DeleteFirewallRule(r.Context(), id)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "DELETE_FIREWALL_RULE", idStr, status, "", r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Firewall rule deleted successfully."})
}

type ToggleFirewallRequest struct {
	Enable  bool `json:"enable"`
	SSHPort int  `json:"ssh_port"`
}

// ToggleFirewall handles POST /api/v1/security/firewall/toggle
func (h *SecurityHandler) ToggleFirewall(w http.ResponseWriter, r *http.Request) {
	var req ToggleFirewallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	err := security.ToggleFirewall(r.Context(), req.Enable, req.SSHPort)
	actionStr := "ENABLE_FIREWALL"
	if !req.Enable {
		actionStr = "DISABLE_FIREWALL"
	}
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", actionStr, "UFW", status, "", r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	stateStr := "enabled"
	if !req.Enable {
		stateStr = "disabled"
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Firewall " + stateStr + " successfully."})
}

// ListUsers handles GET /api/v1/security/users
func (h *SecurityHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := security.ListLinuxUsers(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list users"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Shell    string `json:"shell"`
	IsSudo   bool   `json:"is_sudo"`
}

// CreateUser handles POST /api/v1/security/users
func (h *SecurityHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, `{"error":"username cannot be empty"}`, http.StatusBadRequest)
		return
	}

	err := security.CreateLinuxUser(r.Context(), req.Username, req.Shell, req.IsSudo)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "CREATE_USER", req.Username, status, "", r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("User %s created successfully.", req.Username),
	})
}

// DeleteUser handles DELETE /api/v1/security/users/{username}
func (h *SecurityHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, `{"error":"username cannot be empty"}`, http.StatusBadRequest)
		return
	}

	err := security.DeleteLinuxUser(r.Context(), username)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "DELETE_USER", username, status, "", r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("User %s deleted successfully.", username),
	})
}

