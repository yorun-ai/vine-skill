package impl

import (
	skeled "example.com/greeting/skeled/golang"
	"example.com/greeting/src/server/core"
)

// GreetingService adapts the generated Server. Default style (no ER): methods
// return values only; business errors are expressed with ex.PanicNew in core and
// recovered by the framework into structured error responses.
type GreetingService struct {
	skeled.DefaultGreetingServiceServer
	Service *core.GreetingService `inject:""`
}

func (h *GreetingService) CreateGreeting(message string) skeled.Greeting {
	return h.Service.Create(message)
}

func (h *GreetingService) ListGreetings(pageToken *string, pageSize *int) skeled.GreetingPage {
	return h.Service.List(pageToken, pageSize)
}
