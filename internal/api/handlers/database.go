package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/codingguna/aio-panel/internal/config"
	"github.com/codingguna/aio-panel/internal/database"
	"github.com/codingguna/aio-panel/internal/db"
)

type DatabaseHandler struct {
	cfg   *config.Config
	store *db.Store
}

func NewDatabaseHandler(cfg *config.Config, store *db.Store) *DatabaseHandler {
	return &DatabaseHandler{cfg: cfg, store: store}
}

// ListPostgresDBs handles GET /api/v1/databases/postgres
func (h *DatabaseHandler) ListPostgresDBs(w http.ResponseWriter, r *http.Request) {
	dbs, err := database.ListPostgresDatabases(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list postgres databases"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dbs)
}

// ListPostgresUsers handles GET /api/v1/databases/postgres/users
func (h *DatabaseHandler) ListPostgresUsers(w http.ResponseWriter, r *http.Request) {
	users, err := database.ListPostgresUsers(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list postgres users"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

type BackupDBRequest struct {
	Database string `json:"database"`
}

// BackupPostgres handles POST /api/v1/databases/postgres/backup
func (h *DatabaseHandler) BackupPostgres(w http.ResponseWriter, r *http.Request) {
	var req BackupDBRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	backupFile, err := database.BackupPostgresDatabase(r.Context(), req.Database, h.cfg.Paths.BackupDir)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "BACKUP_POSTGRES", req.Database, status, backupFile, r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"backup_file": backupFile,
		"message":     "PostgreSQL backup created successfully.",
	})
}

// ListMySQLDBs handles GET /api/v1/databases/mysql
func (h *DatabaseHandler) ListMySQLDBs(w http.ResponseWriter, r *http.Request) {
	dbs, err := database.ListMySQLDatabases(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list mysql databases"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dbs)
}

// BackupMySQL handles POST /api/v1/databases/mysql/backup
func (h *DatabaseHandler) BackupMySQL(w http.ResponseWriter, r *http.Request) {
	var req BackupDBRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	backupFile, err := database.BackupMySQLDatabase(r.Context(), req.Database, h.cfg.Paths.BackupDir)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "BACKUP_MYSQL", req.Database, status, backupFile, r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"backup_file": backupFile,
		"message":     "MySQL backup created successfully.",
	})
}

type CreateDBRequest struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

// CreatePostgresDB handles POST /api/v1/databases/postgres/create
func (h *DatabaseHandler) CreatePostgresDB(w http.ResponseWriter, r *http.Request) {
	var req CreateDBRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	err := database.CreatePostgresDatabase(r.Context(), req.Name, req.Owner)
	status := "SUCCESS"
	var errDetails string
	if err != nil {
		status = "FAILURE"
		errDetails = err.Error()
	}

	if h.store != nil {
		_ = h.store.LogAudit(r.Context(), "admin", "CREATE_POSTGRES_DB", req.Name, status, errDetails, r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message":  "PostgreSQL database created successfully: " + req.Name,
		"database": req.Name,
	})
}

// CreateMySQLDB handles POST /api/v1/databases/mysql/create
func (h *DatabaseHandler) CreateMySQLDB(w http.ResponseWriter, r *http.Request) {
	var req CreateDBRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	err := database.CreateMySQLDatabase(r.Context(), req.Name)
	status := "SUCCESS"
	var errDetails string
	if err != nil {
		status = "FAILURE"
		errDetails = err.Error()
	}

	if h.store != nil {
		_ = h.store.LogAudit(r.Context(), "admin", "CREATE_MYSQL_DB", req.Name, status, errDetails, r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message":  "MySQL database created successfully: " + req.Name,
		"database": req.Name,
	})
}

// DeletePostgresDB handles DELETE /api/v1/databases/postgres/{name}
func (h *DatabaseHandler) DeletePostgresDB(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, `{"error":"missing database name"}`, http.StatusBadRequest)
		return
	}

	err := database.DeletePostgresDatabase(r.Context(), name)
	status := "SUCCESS"
	var errDetails string
	if err != nil {
		status = "FAILURE"
		errDetails = err.Error()
	}

	if h.store != nil {
		_ = h.store.LogAudit(r.Context(), "admin", "DELETE_POSTGRES_DB", name, status, errDetails, r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message": "PostgreSQL database " + name + " dropped successfully",
	})
}

// DeleteMySQLDB handles DELETE /api/v1/databases/mysql/{name}
func (h *DatabaseHandler) DeleteMySQLDB(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, `{"error":"missing database name"}`, http.StatusBadRequest)
		return
	}

	err := database.DeleteMySQLDatabase(r.Context(), name)
	status := "SUCCESS"
	var errDetails string
	if err != nil {
		status = "FAILURE"
		errDetails = err.Error()
	}

	if h.store != nil {
		_ = h.store.LogAudit(r.Context(), "admin", "DELETE_MYSQL_DB", name, status, errDetails, r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message": "MySQL database " + name + " dropped successfully",
	})
}
