package repository

import (
	"errors"

	"goAuth/internals/model"
)

var (
	ErrNotFound = errors.New("user not found")
	ErrConflict = errors.New("user already exists")
)

type UserRepository interface {
	Create(user *model.User) error
	GetByID(id string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetByEmailOrPhone(emailOrPhone string) (*model.User, error)
}
