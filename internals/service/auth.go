package service

import (
	"context"
	"goAuth/internals/model"
)

type AuthService interface {
	Signup(ctx context.Context, req *model.SignupRequest) (*model.User, error)
	Login(ctx context.Context, loginID, password string, useOTP bool) (*model.User, *model.TokenResponse, error)
	VerifyOTP(ctx context.Context, emailOrPhone, otp string) (*model.User, *model.TokenResponse, error)
	GetByID(ctx context.Context, userID string) (*model.User, error)
}
