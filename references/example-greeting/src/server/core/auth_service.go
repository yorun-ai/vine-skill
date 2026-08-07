package core

import (
	skeled "example.com/greeting/skeled/golang"
)

// Demo credentials. Replace with real authentication in a production project.
const (
	DefaultUsername = "admin"
	DefaultPassword = "admin123"
)

type AuthService struct {
	Sessions *SessionStore `inject:""`
}

// Login returns success=false on mismatch; on success it issues a session token
// that the caller must send as the actor credential (Authorization header).
func (s *AuthService) Login(username, password string) skeled.LoginResult {
	if username != DefaultUsername || password != DefaultPassword {
		return skeled.LoginResult{Success: false}
	}
	token := s.Sessions.Issue(username)
	return skeled.LoginResult{Success: true, Username: &username, Token: &token}
}
