package middlewares

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"

	"goAuth/internals/helpers"
	"goAuth/internals/types"
)

var (
	globalJWT    helpers.TokenProvider
	globalJWTOnce sync.Once
)

// GetGlobalJWT returns the JWT provider from env (JWT_SECRET). Used for token validation on protected routes.
func GetGlobalJWT() helpers.TokenProvider {
	globalJWTOnce.Do(func() {
		secret := os.Getenv(types.EnvJWTSecret)
		if secret == "" {
			secret = "change-me-in-production"
		}
		globalJWT = helpers.NewJWTProvider(secret, types.TokenExpiry)
	})
	return globalJWT
}

type contextKey string

const ClaimsContextKey contextKey = "claims"

func Auth(tokenProvider helpers.TokenProvider) func(http.Handler) http.Handler {
	return AuthFromContext(func(context.Context) helpers.TokenProvider { return tokenProvider })
}

func AuthFromContext(getJWT func(context.Context) helpers.TokenProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenProvider := getJWT(r.Context())
			if tokenProvider == nil {
				helpers.WriteError(w, http.StatusUnauthorized, types.MsgUnauthorized)
				return
			}
			auth := r.Header.Get("Authorization")
			if auth == "" {
				helpers.WriteError(w, http.StatusUnauthorized, types.MsgUnauthorized)
				return
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				helpers.WriteError(w, http.StatusUnauthorized, types.MsgUnauthorized)
				return
			}
			tokenString := strings.TrimSpace(parts[1])
			if tokenString == "" {
				helpers.WriteError(w, http.StatusUnauthorized, types.MsgUnauthorized)
				return
			}
			claims, err := tokenProvider.Validate(tokenString)
			if err != nil {
				helpers.WriteError(w, http.StatusUnauthorized, types.MsgInvalidToken)
				return
			}
			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetClaims(ctx context.Context) *helpers.Claims {
	c, _ := ctx.Value(ClaimsContextKey).(*helpers.Claims)
	return c
}
