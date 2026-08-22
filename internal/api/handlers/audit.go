package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/codingguna/aio-panel/internal/db"
)

type AuditHandler struct {
	store *db.Store
}

func NewAuditHandler(store *db.Store) *AuditHandler {
	return &AuditHandler{store: store}
}

// GetAuditEvents handles GET /api/v1/audit/events
func (h *AuditHandler) GetAuditEvents(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		http.Error(w, `{"error":"database unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	events, err := h.store.GetRecentAuditEvents(r.Context(), limit)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch audit events"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
