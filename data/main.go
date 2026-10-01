package data

import (
	"sync"
	"time"
)

type Status string

const (
	StatusOrderPending      Status = "pending"
	StatusOrderProcessing   Status = "processing"
	StatusOrderVoided       Status = "voided"
	StatusOrderCompleted    Status = "completed"
	StatusPaymentDeclined   Status = "payment_declined"
	StatusPaymentAuthorized Status = "payment_authorized"
	StatusNeedsAttention    Status = "needs_attention"
)

type Event struct {
	Status Status    `json:"status"`
	At     time.Time `json:"at"`
}

type Order struct {
	ID              int       `json:"id"`
	UserID          string    `json:"user_id"`
	Amount          int       `json:"amount"`
	Status          Status    `json:"status"`
	StatusUpdatedAt time.Time `json:"status_updated_at"`
	History         []Event   `json:"history"`
}

type Store struct {
	mu          sync.Mutex
	DataMap     map[int]Order
	IdIncrement int
}

func NewStore() *Store {
	return &Store{
		DataMap: make(map[int]Order),
	}
}

func (s *Store) Get(orderID int) (Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.DataMap[orderID]
	if !ok {
		return Order{}, ErrNotFound
	}

	return order, nil
}

func (s *Store) SetStatusPending(o Order) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.IdIncrement + 1
	s.IdIncrement = id

	o.ID = id
	s.DataMap[id] = withStatus(o, StatusOrderPending)

	return o.ID
}

// ClaimPending moves a pending order to processing in one locked step, so two
// concurrent requests can't both see "pending" and both authorize payment.
func (s *Store) ClaimPending(orderID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.DataMap[orderID]
	if !ok {
		return ErrNotFound
	}
	if order.Status != StatusOrderPending {
		return ErrInvalidTransition
	}

	s.DataMap[orderID] = withStatus(order, StatusOrderProcessing)
	return nil
}

func (s *Store) SetStatusAuthorized(orderID int) {
	s.setStatus(orderID, StatusPaymentAuthorized)
}

func (s *Store) SetStatusDeclined(orderID int) {
	s.setStatus(orderID, StatusPaymentDeclined)
}

func (s *Store) SetStatusVoided(orderID int) {
	s.setStatus(orderID, StatusOrderVoided)
}

func (s *Store) SetStatusComplete(orderID int) {
	s.setStatus(orderID, StatusOrderCompleted)
}

func (s *Store) SetStatusNeedsAttention(orderID int) {
	s.setStatus(orderID, StatusNeedsAttention)
}

// setStatus updates an existing order. A missing ID is ignored rather than
// stored as a zero-value order.
func (s *Store) setStatus(orderID int, status Status) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.DataMap[orderID]
	if !ok {
		return
	}

	s.DataMap[orderID] = withStatus(order, status)
}

// withStatus sets the status and records it in the order's history. History
// gets a fresh slice so copies of the order returned by Get never share it.
func withStatus(o Order, status Status) Order {
	now := time.Now()
	o.Status = status
	o.StatusUpdatedAt = now
	o.History = append(o.History[:len(o.History):len(o.History)], Event{Status: status, At: now})
	return o
}
