package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/codingguna/aio-panel/internal/version"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
}

var startTime = time.Now()

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := HealthResponse{
		Status:    "healthy",
		Version:   version.Version,
		Timestamp: time.Now().UTC(),
		Uptime:    time.Since(startTime).Round(time.Second).String(),
	}
	json.NewEncoder(w).Encode(resp)
}
