package types

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("user already exists")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidToken    = errors.New("invalid or expired token")
	ErrInvalidOTP      = errors.New("invalid or expired OTP")
	ErrOTPSent         = errors.New("otp sent") // Login with use_otp: OTP emailed, no token yet
	ErrValidation      = errors.New("validation failed")
	ErrBcryptCost      = errors.New("bcrypt cost must be between 4 and 31")
)
