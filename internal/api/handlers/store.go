package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/codingguna/aio-panel/internal/db"
	"github.com/codingguna/aio-panel/internal/ops"
)

type StoreHandler struct {
	store     *db.Store
	installer *ops.PackageInstaller
}

func NewStoreHandler(store *db.Store) *StoreHandler {
	return &StoreHandler{
		store:     store,
		installer: ops.GetPackageInstaller(),
	}
}

// ListPackages handles GET /api/v1/store/packages
func (h *StoreHandler) ListPackages(w http.ResponseWriter, r *http.Request) {
	packages := h.installer.GetCatalog(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(packages)
}

// SearchPackages handles GET /api/v1/store/search?q=query
func (h *StoreHandler) SearchPackages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	packages, err := h.installer.SearchSystemPackages(r.Context(), q)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(packages)
}

type InstallPackageRequest struct {
	PackageID string `json:"package_id"`
	Version   string `json:"version"`
}

// InstallPackage handles POST /api/v1/store/install
func (h *StoreHandler) InstallPackage(w http.ResponseWriter, r *http.Request) {
	var req InstallPackageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	job, err := h.installer.InstallPackage(r.Context(), req.PackageID, req.Version)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if h.store != nil {
		_ = h.store.LogAudit(r.Context(), "admin", "INSTALL_PACKAGE", req.PackageID, "SUCCESS", "Initiated installation: "+req.Version, r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// GetJobStatus handles GET /api/v1/store/jobs/{id}
func (h *StoreHandler) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job := h.installer.GetJobStatus(id)
	if job == nil {
		http.Error(w, `{"error":"job not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}
