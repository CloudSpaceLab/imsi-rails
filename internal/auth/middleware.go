package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

var ErrForbidden = errors.New("forbidden")

type contextKey string

const userContextKey contextKey = "imsi-auth-user"

type Authorizer func(Permission, http.Handler) http.Handler

func (s *Service) Require(permission Permission, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := s.SessionFromRequest(r)
		if !ok {
			writeProblem(w, http.StatusUnauthorized, "unauthenticated", "A valid session is required.")
			return
		}
		if !session.User.HasPermission(permission) {
			writeProblem(w, http.StatusForbidden, "forbidden", "This action is not allowed for your role.")
			return
		}
		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), session.User)))
	})
}

func WithUser(ctx context.Context, user SessionUser) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (SessionUser, bool) {
	user, ok := ctx.Value(userContextKey).(SessionUser)
	return user, ok
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeProblem(w http.ResponseWriter, status int, code, detail string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":   code,
			"detail": detail,
		},
	})
}
