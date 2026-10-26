package catalog

import (
	skeled "example.com/shop/skeled/catalog"
	orderspub "example.com/shop/skeled/orderspub"
)

type OrderPlacedListener struct {
	orderspub.DefaultOrderPlacedEventListener
	Orders       orderspub.OrderServiceClient          `inject:""`
	Reservations *Reservations                         `inject:""`
	Events       skeled.ReservationCreatedEventEmitter `inject:""`
}

func (h *OrderPlacedListener) OnOrderPlaced(event *orderspub.OrderPlacedEvent) {
	// Reverse-direction Rpc: the catalog reads the order through its public client.
	order := h.Orders.GetOrder(event.OrderId)
	reservation := h.Reservations.Record(order.Id, order.Product.Title, order.Quantity)
	// Retrying publication reuses the recorded event and its stable identity.
	h.Events.EmitReservationCreated(&reservation)
}
