package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ruhan/internal/store"
	"github.com/ruhan/internal/tokens"
	"github.com/ruhan/internal/utils"
)

type UseMiddleware struct {
	UserStore store.UserStore
}

type contextKey string

const USER_CONTEXT_KEY = contextKey("user")

func SetUser(req *http.Request, user *store.User) *http.Request {
	ctx := context.WithValue(req.Context(), USER_CONTEXT_KEY, user)
	return req.WithContext(ctx)
}

func GetUser(req *http.Request) *store.User {
	user, ok := req.Context().Value(USER_CONTEXT_KEY).(*store.User)
	if !ok {
		panic("missing user in request")
	}
	return user
}

func (u *UseMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Vary", "Authorization")
		authHeader := res.Header().Get("Authorization")
		if authHeader == "" {
			r := SetUser(req, store.AnonymousUser)
			next.ServeHTTP(res, r)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			utils.WriteJSON(res, http.StatusUnauthorized, utils.Envelope{"error": "invalid authorization header"})
			return
		}

		token := headerParts[1]
		user, err := u.UserStore.GetUserToken(tokens.ScopeAuth, token)
		if err != nil || user == nil {
			utils.WriteJSON(res, http.StatusUnauthorized, utils.Envelope{"error": "invalid token"})
			return
		}

		r := SetUser(req, user)
		next.ServeHTTP(res, r)
	})
}

func (u *UseMiddleware) RequireUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		user := GetUser(req)

		if user.IsAnonymous() {
			utils.WriteJSON(res, http.StatusUnauthorized, utils.Envelope{"error": "you must be logged in to access this route"})
			return
		}

		next.ServeHTTP(res, req)
	})
}
