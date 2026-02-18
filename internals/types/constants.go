package types

import "time"

var (
	TokenExpiry = 24 * time.Hour
)

const (
	EnvConfigPath    = "CONFIG_PATH"
	EnvDatabaseDSN   = "DATABASE_DSN"
	EnvJWTSecret     = "JWT_SECRET"
	EnvPort          = "PORT"
	HeaderProjectID  = "X-Project-ID"
	DefaultProjectID = "default"
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
)
