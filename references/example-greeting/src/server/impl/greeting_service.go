package impl

import (
	skeled "example.com/greeting/skeled/golang"
	"example.com/greeting/src/server/core"
)

// GreetingService adapts the generated Server. Methods return business values;
// core expresses business errors with ex.PanicNew, which are
// recovered by the framework into structured error responses.
type GreetingService struct {
	skeled.DefaultGreetingApiServiceServer
	Service *core.GreetingService `inject:""`
}

func (h *GreetingService) CreateGreeting(message string) skeled.Greeting {
	return h.Service.Create(message)
}

func (h *GreetingService) ListGreetings(pageToken *string, pageSize *int) skeled.GreetingPage {
	return h.Service.List(pageToken, pageSize)
}
