package catalog

import (
	"time"

	vineapp "go.yorun.ai/vine/app"
)

type App struct {
	vineapp.Application
	vineapp.ServicerEnabled
	vineapp.EventerEnabled
}

func (*App) Name() string { return "shop.catalog" }

func (*App) InitModules(add vineapp.TypeAdder) {
	add(vineapp.T[*Reservations]())
}

func (*App) ServicerInitHandlers(add vineapp.TypeAdder) {
	add(vineapp.T[*CatalogService]())
}

func (*App) EventerInitListeners(add vineapp.ListenerTypeAdder) {
	add(vineapp.T[*OrderPlacedListener](),
		vineapp.WithListenerTimeout(20*time.Second),
		vineapp.WithListenerConcurrency(2))
}
