package orders

import (
	"sync"
	"uuid"

	skeltype "go.yorun.ai/skel/types"
	vineapp "go.yorun.ai/vine/app"
	"go.yorun.ai/vine/core/ex"

	catalogpub "example.com/shop/skeled/catalogpub"
	skeled "example.com/shop/skeled/orders"
)

type OrderStore struct {
	vineapp.BaseModule
	mu      sync.Mutex
	records map[string]skeled.Order
}

func (s *OrderStore) BeforeAppStart() error {
	s.records = make(map[string]skeled.Order)
	return nil
}

func (s *OrderStore) Create(product catalogpub.Product, quantity int) skeled.Order {
	s.mu.Lock()
	defer s.mu.Unlock()
	order := skeled.Order{
		Id:       skeltype.NewUUID(uuid.NewV7()),
		Product:  product,
		Quantity: quantity,
	}
	s.records[order.Id.String()] = order
	return order
}

func (s *OrderStore) Get(orderId skeltype.UUID) skeled.Order {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.get(orderId)
}

func (s *OrderStore) Confirm(orderId skeltype.UUID, confirmation string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	order := s.get(orderId)
	// Repeating the same notification preserves the completed state.
	order.Confirmed = true
	order.Confirmation = confirmation
	s.records[order.Id.String()] = order
}

func (s *OrderStore) get(orderId skeltype.UUID) skeled.Order {
	order, found := s.records[orderId.String()]
	if !found {
		ex.PanicNew(ex.NotFound, "order not found")
	}
	return order
}
