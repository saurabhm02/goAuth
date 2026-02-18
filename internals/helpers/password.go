package helpers

import (
	"goAuth/internals/types"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// BcryptHasher: only Hash and Verify. Cost is fixed at 12.
type BcryptHasher struct {
	Cost int
}

func (h BcryptHasher) Hash(password string) (string, error) {
	cost := h.Cost
	if cost <= 0 {
		cost = bcryptCost
	}
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return "", types.ErrBcryptCost
	}
	b, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h BcryptHasher) Verify(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// NewBcryptHasher returns a bcrypt hasher with cost 12 (same for all projects).
func NewBcryptHasher() BcryptHasher {
	return BcryptHasher{Cost: bcryptCost}
}
