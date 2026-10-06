package service

import "context"

type PaymentServicer interface {
	Authorize(ctx context.Context, orderID int) error
	Void(ctx context.Context, orderID int) error
}

type Completer interface {
	Complete(ctx context.Context, orderID int) error
}
