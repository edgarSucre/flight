package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/edgarSucre/flight/token"
)

type ctxKey string

const (
	authHeader          = "Authorization"
	bearerPrefix        = "Bearer "
	userCtxKey   ctxKey = "username"
)

func jwtMiddleware(next http.Handler, tokenMaker token.Maker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/user/login" {
			next.ServeHTTP(w, r)
			return
		}

		auth := r.Header.Get(authHeader)
		if len(auth) == 0 {
			err := fmt.Sprintf("missing %s header", authHeader)
			http.Error(w, err, http.StatusUnauthorized)

			return
		}

		if !strings.HasPrefix(auth, bearerPrefix) {
			err := fmt.Sprintf("missing %s token", bearerPrefix)
			http.Error(w, err, http.StatusUnauthorized)
			return
		}

		username, err := tokenMaker.VerifyToken(auth[len(bearerPrefix):])
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userCtxKey, username)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
