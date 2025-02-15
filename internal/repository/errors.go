package repository

import "errors"

var (
	ErrNotFound          = errors.New("record not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidMerch      = errors.New("invalid merch name")
)
