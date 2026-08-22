package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// Session represents an active administrative user session
type Session struct {
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SessionManager handles cryptographic session tokens and in-memory cache
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	// Rate limiting map: IP -> attempts
	rateMu   sync.Mutex
	attempts map[string]*RateLimit
}

type RateLimit struct {
	Count     int
	BlockedAt time.Time
}

func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		attempts: make(map[string]*RateLimit),
	}
	// Background cleanup of expired sessions every 10 minutes
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		for range ticker.C {
			sm.cleanup()
		}
	}()
	return sm
}

// CreateSession generates a secure 32-byte session token
func (sm *SessionManager) CreateSession(username, role string) (*Session, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(bytes)

	sess := &Session{
		Token:     token,
		Username:  username,
		Role:      role,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour), // 24-hour validity
	}

	sm.mu.Lock()
	sm.sessions[token] = sess
	sm.mu.Unlock()

	return sess, nil
}

// ValidateSession verifies if a session token is valid and active
func (sm *SessionManager) ValidateSession(token string) (*Session, bool) {
	if token == "" {
		return nil, false
	}

	sm.mu.RLock()
	sess, exists := sm.sessions[token]
	sm.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().UTC().After(sess.ExpiresAt) {
		sm.RevokeSession(token)
		return nil, false
	}

	return sess, true
}

// RevokeSession invalidates a session token
func (sm *SessionManager) RevokeSession(token string) {
	sm.mu.Lock()
	delete(sm.sessions, token)
	sm.mu.Unlock()
}

// CheckRateLimit checks if an IP is temporarily blocked (max 5 failed attempts in 5 minutes)
func (sm *SessionManager) CheckRateLimit(ip string) bool {
	sm.rateMu.Lock()
	defer sm.rateMu.Unlock()

	rl, exists := sm.attempts[ip]
	if !exists {
		return true
	}

	if time.Since(rl.BlockedAt) > 5*time.Minute {
		delete(sm.attempts, ip)
		return true
	}

	return rl.Count < 5
}

// RecordFailedAttempt registers a failed login attempt for an IP
func (sm *SessionManager) RecordFailedAttempt(ip string) {
	sm.rateMu.Lock()
	defer sm.rateMu.Unlock()

	rl, exists := sm.attempts[ip]
	if !exists {
		sm.attempts[ip] = &RateLimit{Count: 1, BlockedAt: time.Now()}
		return
	}

	rl.Count++
	rl.BlockedAt = time.Now()
}

// ResetRateLimit clears failed attempts on successful login
func (sm *SessionManager) ResetRateLimit(ip string) {
	sm.rateMu.Lock()
	delete(sm.attempts, ip)
	sm.rateMu.Unlock()
}

func (sm *SessionManager) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now().UTC()
	for token, sess := range sm.sessions {
		if now.After(sess.ExpiresAt) {
			delete(sm.sessions, token)
		}
	}
}
