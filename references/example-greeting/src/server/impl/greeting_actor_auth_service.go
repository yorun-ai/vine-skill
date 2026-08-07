package impl

import (
	skeled "example.com/greeting/skeled/golang"
	"example.com/greeting/src/server/core"
)

// GreetingActorAuthService validates the actor credential (session token).
// Portal calls it during admission before delivering an authenticated actor.
type GreetingActorAuthService struct {
	skeled.DefaultGreetingActorAuthServiceServer
	Service *core.GreetingActorAuthService `inject:""`
}

func (h *GreetingActorAuthService) Auth(credential skeled.GreetingActorCredential) skeled.GreetingActorInfo {
	return h.Service.Auth(credential)
}
