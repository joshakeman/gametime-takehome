package service

import "context"

// FakePaymentService stands in for a real payment provider. The zero value
// succeeds; set a flag to force a failure. Like a real client, it gives up if
// the context is already cancelled or past its deadline.
type FakePaymentService struct {
	ShouldDecline  bool
	ShouldFailVoid bool
}

func (ps FakePaymentService) Authorize(ctx context.Context, orderID int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if ps.ShouldDecline {
		return ErrPaymentDeclined
	}
	return nil
}

func (ps FakePaymentService) Void(ctx context.Context, orderID int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if ps.ShouldFailVoid {
		return ErrVoidFailed
	}
	return nil
}

// FakeCompleter stands in for order fulfillment. The zero value succeeds.
type FakeCompleter struct {
	ShouldFailComplete bool
}

func (c FakeCompleter) Complete(ctx context.Context, orderID int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if c.ShouldFailComplete {
		return ErrCompleteFailed
	}
	return nil
}
