package core

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

// SessionStore keeps issued session tokens in memory (demo only). A production
// project would persist sessions in a database or issue signed tokens.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]string // token -> username
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: map[string]string{}}
}

func (s *SessionStore) Issue(username string) string {
	token := randomToken()
	s.mu.Lock()
	s.sessions[token] = username
	s.mu.Unlock()
	return token
}

// Validate returns the username for a valid session token.
func (s *SessionStore) Validate(token string) (string, bool) {
	s.mu.RLock()
	username, ok := s.sessions[token]
	s.mu.RUnlock()
	return username, ok
}

func randomToken() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
