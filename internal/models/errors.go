package models

import "errors"

var (
	ErrInsufficientFunds = errors.New("insufficient funds on balance")
)