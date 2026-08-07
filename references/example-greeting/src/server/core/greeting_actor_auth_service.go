package core

import (
	"go.yorun.ai/vine/core/ex"

	skeled "example.com/greeting/skeled/golang"
)

// GreetingActorAuthService validates the actor credential (session token) for
// Portal. Portal calls this before treating a request as authenticated.
type GreetingActorAuthService struct {
	Sessions *SessionStore `inject:""`
}

func (s *GreetingActorAuthService) Auth(credential skeled.GreetingActorCredential) skeled.GreetingActorInfo {
	username, ok := s.Sessions.Validate(credential.Token)
	if !ok {
		ex.PanicNew(ex.Unauthorized, "invalid session token")
	}
	return skeled.GreetingActorInfo{Username: username}
}
