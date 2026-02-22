package helpers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const (
	OTPLength    = 6
	OTPExpiryMin = 5
)

var OTPExpiry = OTPExpiryMin * time.Minute

func GenerateOTP() (code string, hash string, err error) {
	var digits string
	for i := 0; i < OTPLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", "", err
		}
		digits += fmt.Sprintf("%d", n.Int64())
	}
	// Save as plain text
	return digits, digits, nil
}

func VerifyOTP(code, hash string) bool {
	return code == hash
}

func OTPExpiryTime() time.Time {
	return time.Now().Add(OTPExpiry)
}
