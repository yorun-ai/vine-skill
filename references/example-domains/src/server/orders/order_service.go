package orders

import (
	skeltype "go.yorun.ai/skel/types"
	"go.yorun.ai/vine/core/ex"

	catalogpub "example.com/shop/skeled/catalogpub"
	skeled "example.com/shop/skeled/orders"
)

type OrderApiService struct {
	skeled.DefaultOrderApiServiceServer
	Catalog catalogpub.CatalogServiceClient `inject:""`
	Store   *OrderStore                     `inject:""`
	Events  skeled.OrderPlacedEventEmitter  `inject:""`
}

func (h *OrderApiService) PlaceOrder(sku string, quantity int) skeled.Order {
	if quantity < 1 || quantity > 100 {
		ex.PanicNew(ex.ValidationFailed, "quantity must be between 1 and 100")
	}
	product := h.Catalog.GetProduct(sku)
	order := h.Store.Create(product, quantity)
	h.Events.EmitOrderPlaced(&skeled.OrderPlacedEvent{EventId: order.Id, OrderId: order.Id})
	return order
}

func (h *OrderApiService) GetOrder(orderId skeltype.UUID) skeled.Order {
	return h.Store.Get(orderId)
}

type OrderService struct {
	skeled.DefaultOrderServiceServer
	Store *OrderStore `inject:""`
}

func (h *OrderService) GetOrder(orderId skeltype.UUID) skeled.Order {
	return h.Store.Get(orderId)
}
