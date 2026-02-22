package middlewares

import (
	"context"
	"net/http"

	"goAuth/internals/helpers"
	"goAuth/internals/logger"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = helpers.GenerateID()
		}

		ctx := context.WithValue(r.Context(), logger.RequestIDKey, reqID)

		w.Header().Set("X-Request-ID", reqID)

		logger.Logf(ctx, "middleware", "-", "Incoming request: %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
