package catalog

import (
	"fmt"
	"sync"

	skeltype "go.yorun.ai/skel/types"
	vineapp "go.yorun.ai/vine/app"

	skeled "example.com/shop/skeled/catalog"
)

// Reservations owns demo state for this App's lifetime.
type Reservations struct {
	vineapp.BaseModule
	mu      sync.Mutex
	records map[string]skeled.ReservationCreatedEvent
}

func (r *Reservations) BeforeAppStart() error {
	r.records = make(map[string]skeled.ReservationCreatedEvent)
	return nil
}

func (r *Reservations) Record(orderId skeltype.UUID, title string, quantity int) skeled.ReservationCreatedEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	if previous, found := r.records[orderId.String()]; found {
		return previous
	}
	reservation := skeled.ReservationCreatedEvent{
		EventId:      orderId,
		OrderId:      orderId,
		Confirmation: fmt.Sprintf("Reserved %d × %s", quantity, title),
	}
	r.records[orderId.String()] = reservation
	return reservation
}
