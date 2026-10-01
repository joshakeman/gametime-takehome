package service

type PaymentServicer interface {
	Authorize(orderID int) error
	Void(orderID int) error
}

type Completer interface {
	Complete(orderID int) error
}
