package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	DB     *sql.DB
	Secret string
}

type contextKey string

// UserIDKey is the context key used by jwtMiddleware to inject the authenticated user ID.
const UserIDKey contextKey = "user_id"

func userIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"code": code, "message": message})
}
