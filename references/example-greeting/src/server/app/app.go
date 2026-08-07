package app

import (
	vineapp "go.yorun.ai/vine/app"
	"go.yorun.ai/vine/core/di"

	"example.com/greeting/src/server/core"
	"example.com/greeting/src/server/impl"
	"example.com/greeting/src/server/repo"
)

type GreetingApp struct {
	vineapp.Application
	vineapp.ServicerEnabled
}

func (*GreetingApp) Name() string {
	return "greeting"
}

func (*GreetingApp) InitComponents(add vineapp.TypeAdder) {
	add(vineapp.T[*MainDatabase]())
}

func (*GreetingApp) InitModules(add vineapp.TypeAdder) {
	add(vineapp.T[*Migrator]())
}

func (*GreetingApp) BindCommon(b *di.Binder) {
	b.Bind(di.T[core.GreetingRepository]()).ToImplementation(di.T[*repo.GreetingRepositoryImpl]())
	b.BindInstance(core.NewSessionStore())
}

func (*GreetingApp) ServicerInitHandlers(add vineapp.TypeAdder) {
	add(vineapp.T[*impl.AuthService]())
	add(vineapp.T[*impl.GreetingService]())
	add(vineapp.T[*impl.GreetingActorAuthService]())
	add(vineapp.T[*impl.WelcomeService]())
}
