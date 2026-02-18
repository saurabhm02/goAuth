package types

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("user already exists")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidToken    = errors.New("invalid or expired token")
	ErrValidation      = errors.New("validation failed")
	ErrBcryptCost      = errors.New("bcrypt cost must be between 4 and 31")
)
