package middlewares

import (
	"context"
	"net/http"
	"strings"

	"goAuth/internals/helpers"
	"goAuth/internals/types"
)

const ProjectContextKey contextKey = "project"

func Project(store map[string]interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			projectID := strings.TrimSpace(r.Header.Get(types.HeaderProjectID))
			if projectID == "" {
				projectID = types.DefaultProjectID
			}
			pc, ok := store[projectID]
			if !ok || pc == nil {
				helpers.WriteError(w, http.StatusNotFound, types.MsgProjectNotFound)
				return
			}
			ctxWith := context.WithValue(r.Context(), ProjectContextKey, pc)
			next.ServeHTTP(w, r.WithContext(ctxWith))
		})
	}
}

func GetProjectContext(ctx context.Context) interface{} {
	return ctx.Value(ProjectContextKey)
}
