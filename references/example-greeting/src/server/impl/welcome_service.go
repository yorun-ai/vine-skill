package impl

import (
	"go.yorun.ai/vine/core/meta"

	skeled "example.com/greeting/skeled/golang"
	"example.com/greeting/src/server/core"
)

// WelcomeService reads the authenticated actor (the "session") from the
// execution context. It can only succeed when Portal delivered an
// authenticated actor after validating the session token.
type WelcomeService struct {
	skeled.DefaultWelcomeServiceServer

	Context meta.Context         `inject:""`
	Service *core.WelcomeService `inject:""`
}

func (h *WelcomeService) GetWelcome() skeled.Welcome {
	return h.Service.GetWelcome(h.Context)
}
