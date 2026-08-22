package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/codingguna/aio-panel/internal/config"
	"github.com/codingguna/aio-panel/internal/discovery"
)

type DiscoveryHandler struct {
	cfg *config.Config
}

func NewDiscoveryHandler(cfg *config.Config) *DiscoveryHandler {
	return &DiscoveryHandler{cfg: cfg}
}

// GetPreflightReport handles GET /api/v1/discovery
func (h *DiscoveryHandler) GetPreflightReport(w http.ResponseWriter, r *http.Request) {
	report, err := discovery.RunPreflight(r.Context(), h.cfg)
	if err != nil {
		http.Error(w, `{"error":"failed to generate preflight discovery report"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetApplications handles GET /api/v1/applications
func (h *DiscoveryHandler) GetApplications(w http.ResponseWriter, r *http.Request) {
	apps, err := discovery.DiscoverApplications(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to discover applications"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps)
}

// GetNginxSites handles GET /api/v1/nginx/sites
func (h *DiscoveryHandler) GetNginxSites(w http.ResponseWriter, r *http.Request) {
	sites, err := discovery.DiscoverNginxSites(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to discover nginx sites"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sites)
}
