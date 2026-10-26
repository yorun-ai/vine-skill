package orders

import (
	"time"

	vineapp "go.yorun.ai/vine/app"
)

type App struct {
	vineapp.Application
	vineapp.ServicerEnabled
	vineapp.EventerEnabled
}

func (*App) Name() string { return "shop.orders" }

func (*App) InitModules(add vineapp.TypeAdder) {
	add(vineapp.T[*OrderStore]())
}

func (*App) ServicerInitHandlers(add vineapp.TypeAdder) {
	add(vineapp.T[*OrderApiService]())
	add(vineapp.T[*OrderService]())
}

func (*App) EventerInitListeners(add vineapp.ListenerTypeAdder) {
	add(vineapp.T[*ReservationCreatedListener](),
		vineapp.WithListenerTimeout(20*time.Second),
		vineapp.WithListenerConcurrency(2))
}
