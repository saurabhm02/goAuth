package routes

import (
	"context"
	"net/http"

	"goAuth/internals/handler"
	"goAuth/internals/helpers"
	"goAuth/internals/middlewares"

	"github.com/gorilla/mux"
)

func NewRouter(authHandler *handler.AuthHandler, projectStore map[string]interface{}, authMiddleware func(http.Handler) http.Handler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods(http.MethodGet)

	auth := r.PathPrefix("/auth").Subrouter()
	auth.Use(middlewares.Project(projectStore))
	auth.HandleFunc("/signup", authHandler.Signup).Methods(http.MethodPost)
	auth.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)
	auth.Handle("/me", authMiddleware(http.HandlerFunc(authHandler.Me))).Methods(http.MethodGet)

	return r
}

func AuthMiddlewareFromProject() func(http.Handler) http.Handler {
	return middlewares.AuthFromContext(func(context.Context) helpers.TokenProvider {
		return middlewares.GetGlobalJWT()
	})
}
