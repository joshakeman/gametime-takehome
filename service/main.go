package service

import (
	"errors"

	"github.com/joshakeman/gametime-take-home/data"
)

type OrderService struct {
	store *data.Store
	ps    PaymentServicer
	c     Completer
}

func New(store *data.Store, ps PaymentServicer, c Completer) *OrderService {
	return &OrderService{
		store: store,
		ps:    ps,
		c:     c,
	}
}

func (s *OrderService) Create(userID string, amount int) (data.Order, error) {
	id := s.store.SetStatusPending(data.Order{UserID: userID, Amount: amount})
	return s.store.Get(id)
}

func (s *OrderService) Get(id int) (data.Order, error) {
	return s.store.Get(id)
}

func (s *OrderService) Process(id int) (data.Order, error) {
	// Claim the order atomically: ErrNotFound -> 404, already processed -> 409.
	if err := s.store.ClaimPending(id); err != nil {
		return data.Order{}, err
	}

	if err := s.ps.Authorize(id); err != nil {
		s.store.SetStatusDeclined(id)
		return s.final(id, err)
	}
	s.store.SetStatusAuthorized(id)

	if completeErr := s.c.Complete(id); completeErr != nil {
		if voidErr := s.ps.Void(id); voidErr != nil {
			s.store.SetStatusNeedsAttention(id)
			return s.final(id, errors.Join(completeErr, voidErr)) // keep both
		}
		s.store.SetStatusVoided(id)
		return s.final(id, completeErr)
	}
	s.store.SetStatusComplete(id)
	return s.final(id, nil)
}

// final returns the order as it ended up, plus the error that got it there.
func (s *OrderService) final(id int, err error) (data.Order, error) {
	o, _ := s.store.Get(id)
	return o, err
}
