package orders

import catalogpub "example.com/shop/skeled/catalogpub"

type ReservationCreatedListener struct {
	catalogpub.DefaultReservationCreatedEventListener
	Store *OrderStore `inject:""`
}

func (h *ReservationCreatedListener) OnReservationCreated(event *catalogpub.ReservationCreatedEvent) {
	h.Store.Confirm(event.OrderId, event.Confirmation)
}
