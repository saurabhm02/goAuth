package types

import "time"

var (
	TokenExpiry = 24 * time.Hour
)

const (
	EnvConfigPath          = "CONFIG_PATH"
	EnvDatabaseDSN         = "DATABASE_DSN"
	EnvDatabaseSSLRootCert = "DATABASE_SSL_ROOT_CERT"
	EnvJWTSecret      = "JWT_SECRET"
	EnvPort           = "PORT"
	EnvSMTPHost       = "SMTP_HOST"
	EnvSMTPPort       = "SMTP_PORT"
	EnvSMTPUser       = "SMTP_USER"
	EnvSMTPPassword   = "SMTP_PASSWORD"
	EnvSMTPFrom       = "SMTP_FROM"
	HeaderProjectID   = "X-Project-ID"
	DefaultProjectID  = "default"
)

const (
	MsgInvalidRequest  = "invalid request"
	MsgUnauthorized    = "unauthorized"
	MsgNotFound        = "not found"
	MsgConflict        = "user already exists"
	MsgBadRequest      = "bad request"
	MsgInvalidPassword = "invalid password"
	MsgInvalidToken    = "invalid or expired token"
	MsgPasswordTooWeak = "password does not meet requirements"
	MsgProjectNotFound = "project not found"
	MsgInvalidOTP      = "invalid or expired OTP"
)
