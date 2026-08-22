package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/codingguna/aio-panel/internal/config"
	"github.com/codingguna/aio-panel/internal/db"
	"github.com/codingguna/aio-panel/internal/webserver"
)

type WebServerHandler struct {
	cfg   *config.Config
	store *db.Store
}

func NewWebServerHandler(cfg *config.Config, store *db.Store) *WebServerHandler {
	return &WebServerHandler{cfg: cfg, store: store}
}

// ListVHosts handles GET /api/v1/web/vhosts
func (h *WebServerHandler) ListVHosts(w http.ResponseWriter, r *http.Request) {
	vhosts, err := webserver.ListAIOVHosts()
	if err != nil {
		http.Error(w, `{"error":"failed to list virtual hosts"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vhosts)
}

// CreateVHost handles POST /api/v1/web/vhosts
func (h *WebServerHandler) CreateVHost(w http.ResponseWriter, r *http.Request) {
	var cfg webserver.VHostConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if cfg.Domain == "" {
		http.Error(w, `{"error":"domain cannot be empty"}`, http.StatusBadRequest)
		return
	}

	err := webserver.CreateVHost(r.Context(), cfg, h.cfg.Paths.BackupDir)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "CREATE_VHOST", cfg.Domain, status, string(cfg.Type), r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Virtual host for " + cfg.Domain + " created, validated, and loaded successfully.",
	})
}

// DeleteVHost handles DELETE /api/v1/web/vhosts/{domain}
func (h *WebServerHandler) DeleteVHost(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	if domain == "" {
		http.Error(w, `{"error":"missing domain"}`, http.StatusBadRequest)
		return
	}

	err := webserver.DeleteVHost(r.Context(), domain)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "DELETE_VHOST", domain, status, "", r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Virtual host for " + domain + " deleted safely.",
	})
}

// ListCertificates handles GET /api/v1/web/ssl/certificates
func (h *WebServerHandler) ListCertificates(w http.ResponseWriter, r *http.Request) {
	certs, err := webserver.ListCertificates(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list certificates"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(certs)
}

type IssueSSLRequest struct {
	Domain string `json:"domain"`
	Email  string `json:"email"`
}

// IssueCertificate handles POST /api/v1/web/ssl/issue
func (h *WebServerHandler) IssueCertificate(w http.ResponseWriter, r *http.Request) {
	var req IssueSSLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	err := webserver.IssueCertbotCertificate(r.Context(), req.Domain, req.Email)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "ISSUE_SSL_CERTIFICATE", req.Domain, status, "", r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Let's Encrypt TLS certificate issued and configured successfully for " + req.Domain,
	})
}
