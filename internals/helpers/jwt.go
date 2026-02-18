package helpers

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

type TokenProvider interface {
	Issue(userID string) (string, int64, error)
	Validate(tokenString string) (*Claims, error)
}

type JWTProvider struct {
	Secret []byte
	Expiry time.Duration
}

func NewJWTProvider(secret string, expiry time.Duration) *JWTProvider {
	return &JWTProvider{
		Secret: []byte(secret),
		Expiry: expiry,
	}
}

func (p *JWTProvider) Issue(userID string) (string, int64, error) {
	expiry := p.Expiry
	if expiry <= 0 {
		expiry = 24 * time.Hour
	}
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID,
		},
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(p.Secret)
	if err != nil {
		return "", 0, err
	}
	return s, int64(expiry.Seconds()), nil
}

func (p *JWTProvider) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return p.Secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

var _ TokenProvider = (*JWTProvider)(nil)
