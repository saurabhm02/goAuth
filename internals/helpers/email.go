package helpers

import (
	"fmt"
	"os"
	"strconv"

	"goAuth/internals/types"

	gomail "gopkg.in/gomail.v2"
)

func SendOTPEmail(to, otpCode string) error {
	host := os.Getenv(types.EnvSMTPHost)
	portStr := os.Getenv(types.EnvSMTPPort)
	user := os.Getenv(types.EnvSMTPUser)
	pass := os.Getenv(types.EnvSMTPPassword)
	from := os.Getenv(types.EnvSMTPFrom)
	if host == "" || portStr == "" {
		return fmt.Errorf("SMTP_HOST and SMTP_PORT are required to send OTP email")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("SMTP_PORT must be a number: %w", err)
	}
	if from == "" {
		from = user
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Your verification code")
	m.SetBody("text/plain", fmt.Sprintf("Your OTP is: %s. Valid for %d minutes.", otpCode, OTPExpiryMin))

	d := gomail.NewDialer(host, port, user, pass)
	return d.DialAndSend(m)
}
