package types

import (
	"database/sql"
	"time"

	"goAuth/internals/model"
)

type UserRepository interface {
	Create(user *model.User) error
	DeleteByID(id string) error
	GetByID(id string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetByEmailOrPhone(emailOrPhone string) (*model.User, error)
	CreateTx(tx *sql.Tx, user *model.User) error
}

type OTPRepository interface {
	Create(userID, otpHash string, expiry time.Time) error
	GetByUserID(userID string) (otpHash string, expiry time.Time, err error)
	DeleteByUserID(userID string) error
	DeleteExpired() error
	CreateTx(tx *sql.Tx, userID, otpHash string, expiry time.Time) error
}

type Transactor interface {
	BeginTx() (*sql.Tx, error)
}
