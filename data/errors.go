package data

import "errors"

var (
	ErrNotFound          = errors.New("order not found")
	ErrInvalidTransition = errors.New("invalid state transition")
)
