package impl

import (
	skeled "example.com/greeting/skeled/golang"
	"example.com/greeting/src/server/core"
)

type AuthService struct {
	skeled.DefaultAuthServiceServer
	Service *core.AuthService `inject:""`
}

func (h *AuthService) Login(username string, password string) skeled.LoginResult {
	return h.Service.Login(username, password)
}
