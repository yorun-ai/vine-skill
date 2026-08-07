package core

import (
	"go.yorun.ai/vine/core/ex"
	"go.yorun.ai/vine/core/meta"

	skeled "example.com/greeting/skeled/golang"
)

type WelcomeService struct{}

// GetWelcome reads the authenticated actor from the execution context. This is
// the "session check": Portal only delivers an authenticated actor after the
// session token was validated by GreetingActorAuthService.
func (s *WelcomeService) GetWelcome(ctx meta.Context) skeled.Welcome {
	info, ok := meta.GetActorInfo[*skeled.GreetingActorInfo](ctx.Actor())
	if !ok {
		ex.PanicNew(ex.Unauthorized, "not authenticated")
	}
	return skeled.Welcome{
		Username: info.Username,
		Message:  "Welcome back, " + info.Username,
	}
}
