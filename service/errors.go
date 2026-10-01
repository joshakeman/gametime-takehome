package service

import "errors"

var (
	ErrPaymentDeclined = errors.New("payment was declined")
	ErrVoidFailed      = errors.New("payment void failed")
	ErrCompleteFailed  = errors.New("order completion failed")
)
