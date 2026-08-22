package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/codingguna/aio-panel/internal/auth"
	"github.com/codingguna/aio-panel/internal/db"
)

type contextKey string

const UserContextKey contextKey = "aio_user"

// AuthEnforcer validates authentication on all /api/v1/* routes
func AuthEnforcer(store *db.Store, sm *auth.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// 1. Always allow public endpoints
			if path == "/health" ||
				path == "/api/v1/health" ||
				path == "/api/v1/auth/status" ||
				path == "/api/v1/auth/login" ||
				path == "/api/v1/auth/setup" ||
				!strings.HasPrefix(path, "/api/v1/") {
				next.ServeHTTP(w, r)
				return
			}

			// 2. Check if setup is required
			if store != nil {
				count, _ := store.CountPanelUsers(r.Context())
				if count == 0 {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"error":"Authentication required. Initial setup needed.","setup_required":true}`))
					return
				}
			}

			// 3. Extract and validate session token
			token := extractToken(r)
			if token == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"Unauthorized. Please log in."}`))
				return
			}

			if sm == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Session manager unavailable"}`))
				return
			}

			sess, ok := sm.ValidateSession(token)
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"Session expired or invalid. Please log in again."}`))
				return
			}

			// 4. Inject authenticated user into context
			ctx := context.WithValue(r.Context(), UserContextKey, sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	if cookie, err := r.Cookie("aio_session"); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}
