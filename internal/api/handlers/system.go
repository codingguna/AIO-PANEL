package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/codingguna/aio-panel/internal/db"
	"github.com/codingguna/aio-panel/internal/system"
)

type SystemHandler struct {
	store *db.Store
}

func NewSystemHandler(store *db.Store) *SystemHandler {
	return &SystemHandler{store: store}
}

// GetSystemInfo handles GET /api/v1/system/info
func (h *SystemHandler) GetSystemInfo(w http.ResponseWriter, r *http.Request) {
	info, err := system.GetInfo()
	if err != nil {
		http.Error(w, `{"error":"failed to retrieve system info"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// GetSystemMetrics handles GET /api/v1/system/metrics
func (h *SystemHandler) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := system.GetLiveMetrics()
	if err != nil {
		http.Error(w, `{"error":"failed to sample metrics"}`, http.StatusInternalServerError)
		return
	}

	// Persist snapshot to SQLite for historical graphs
	if h.store != nil {
		info, _ := system.GetInfo()
		hostname := "unknown"
		if info != nil {
			hostname = info.Hostname
		}
		_ = h.store.SaveSystemSnapshot(r.Context(), db.SystemSnapshot{
			Hostname: hostname,
			CPUPct:   metrics.CPU.UsagePercent,
			RAMPct:   metrics.Memory.UsagePercent,
			DiskPct:  metrics.Disk.UsagePercent,
			LoadAvg:  metrics.LoadAverage[0],
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetMetricsHistory handles GET /api/v1/system/history
func (h *SystemHandler) GetMetricsHistory(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		http.Error(w, `{"error":"database unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	snapshots, err := h.store.GetRecentSnapshots(r.Context(), 60)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch history"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshots)
}
