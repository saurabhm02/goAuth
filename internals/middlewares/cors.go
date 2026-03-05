package middlewares

import (
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
	"goAuth/internals/types"
)

func NewCORS() func(http.Handler) http.Handler {
	raw := os.Getenv(types.EnvCORSAllowedOrigins)

	var allowedOrigins []string
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			allowedOrigins = append(allowedOrigins, o)
		}
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Project-ID", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	return c.Handler
}
