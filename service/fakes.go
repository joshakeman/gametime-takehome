package service

// FakePaymentService stands in for a real payment provider. The zero value
// succeeds; set a flag to force a failure.
type FakePaymentService struct {
	ShouldDecline  bool
	ShouldFailVoid bool
}

func (ps FakePaymentService) Authorize(orderID int) error {
	if ps.ShouldDecline {
		return ErrPaymentDeclined
	}
	return nil
}

func (ps FakePaymentService) Void(orderID int) error {
	if ps.ShouldFailVoid {
		return ErrVoidFailed
	}
	return nil
}

// FakeCompleter stands in for order fulfillment. The zero value succeeds.
type FakeCompleter struct {
	ShouldFailComplete bool
}

func (c FakeCompleter) Complete(orderID int) error {
	if c.ShouldFailComplete {
		return ErrCompleteFailed
	}
	return nil
}
