package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrConflict          = errors.New("conflict")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidInput      = errors.New("invalid input")
	ErrOwnPlayer         = errors.New("cannot buy your own player")
)
