package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/codingguna/aio-panel/internal/auth"
	"github.com/codingguna/aio-panel/internal/db"
)

type AuthHandler struct {
	store    *db.Store
	sessions *auth.SessionManager
}

func NewAuthHandler(store *db.Store, sessions *auth.SessionManager) *AuthHandler {
	return &AuthHandler{
		store:    store,
		sessions: sessions,
	}
}

type AuthStatusResponse struct {
	Authenticated bool   `json:"authenticated"`
	Username      string `json:"username,omitempty"`
	Role          string `json:"role,omitempty"`
	SetupRequired bool   `json:"setup_required"`
}

// Status handles GET /api/v1/auth/status
func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	count := 0
	if h.store != nil {
		count, _ = h.store.CountPanelUsers(r.Context())
	}

	setupRequired := (count == 0)

	// Check existing session
	token := extractToken(r)
	if token != "" && h.sessions != nil {
		if sess, ok := h.sessions.ValidateSession(token); ok {
			json.NewEncoder(w).Encode(AuthStatusResponse{
				Authenticated: true,
				Username:      sess.Username,
				Role:          sess.Role,
				SetupRequired: false,
			})
			return
		}
	}

	json.NewEncoder(w).Encode(AuthStatusResponse{
		Authenticated: false,
		SetupRequired: setupRequired,
	})
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	clientIP := r.RemoteAddr

	if h.sessions != nil && !h.sessions.CheckRateLimit(clientIP) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Too many failed login attempts. Please wait 5 minutes before trying again.",
		})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if h.store == nil {
		http.Error(w, `{"error":"database not initialized"}`, http.StatusInternalServerError)
		return
	}

	user, err := h.store.AuthenticateUser(r.Context(), req.Username, req.Password)
	if err != nil {
		if h.sessions != nil {
			h.sessions.RecordFailedAttempt(clientIP)
		}
		_ = h.store.LogAudit(r.Context(), req.Username, "LOGIN_FAILED", "panel_auth", "FAILURE", "Invalid credentials", clientIP)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid username or password"})
		return
	}

	// Successful login: reset rate limiter & create session
	if h.sessions != nil {
		h.sessions.ResetRateLimit(clientIP)
	}

	sess, err := h.sessions.CreateSession(user.Username, user.Role)
	if err != nil {
		http.Error(w, `{"error":"failed to create session"}`, http.StatusInternalServerError)
		return
	}

	// Set HTTP-Only Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "aio_session",
		Value:    sess.Token,
		Path:     "/",
		Expires:  sess.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	_ = h.store.LogAudit(r.Context(), user.Username, "LOGIN_SUCCESS", "panel_auth", "SUCCESS", "Logged in via Web UI", clientIP)

	json.NewEncoder(w).Encode(map[string]any{
		"message":  "Login successful",
		"token":    sess.Token,
		"username": user.Username,
		"role":     user.Role,
	})
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token != "" && h.sessions != nil {
		h.sessions.RevokeSession(token)
	}

	// Expire cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "aio_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

type SetupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Setup handles POST /api/v1/auth/setup (Initial superuser creation if 0 users exist)
func (h *AuthHandler) Setup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.store == nil {
		http.Error(w, `{"error":"database not initialized"}`, http.StatusInternalServerError)
		return
	}

	count, _ := h.store.CountPanelUsers(r.Context())
	if count > 0 {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Initial setup already completed. Use CLI to add users."})
		return
	}

	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Password must be at least 6 characters long"})
		return
	}

	user, err := h.store.CreatePanelUser(r.Context(), req.Username, req.Password, "admin")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	_ = h.store.LogAudit(r.Context(), user.Username, "CREATE_INITIAL_SUPERUSER", "panel_auth", "SUCCESS", "Setup wizard completed", r.RemoteAddr)

	// Automatically create session for the new superuser
	sess, _ := h.sessions.CreateSession(user.Username, user.Role)
	if sess != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "aio_session",
			Value:    sess.Token,
			Path:     "/",
			Expires:  sess.ExpiresAt,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message":  "Superuser created successfully",
		"token":    sess.Token,
		"username": user.Username,
		"role":     user.Role,
	})
}

// ListUsers handles GET /api/v1/auth/users
func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		http.Error(w, `{"error":"database not initialized"}`, http.StatusInternalServerError)
		return
	}

	users, err := h.store.ListPanelUsers(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list users"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// DeleteUser handles DELETE /api/v1/auth/users/{id}
func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusBadRequest)
		return
	}

	if err := h.store.DeletePanelUser(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to delete user"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}

func extractToken(r *http.Request) string {
	// 1. From Cookie
	if cookie, err := r.Cookie("aio_session"); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	// 2. From Authorization header
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}
