package middleware

import (
	"context"
	"fmt"
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// within this anonymouse function
		// we can interject any incoming requests to our server

		w.Header().Add("Vary", "Authorization")
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			r = SetUser(r, store.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authHeader, " ") // Bearer <TOKEN>
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "invalid authorization header"})
			return
		}

		token := headerParts[1]
		fmt.Println("token", token)
		user, err := u.UserStore.GetUserToken(tokens.ScopeAuth, token)
		if err != nil {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "invalid token"})
			return
		}

		if user == nil {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "token expired or invalid"})
			return
		}

		r = SetUser(r, user)
		next.ServeHTTP(w, r)
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
