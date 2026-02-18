package service

import "goAuth/internals/model"

type AuthService interface {
	Signup(req *model.SignupRequest) (*model.User, error)
	Login(loginID, password string) (*model.User, *model.TokenResponse, error)
	GetByID(userID string) (*model.User, error)
}
