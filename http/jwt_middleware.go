package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/edgarSucre/flight"
	"github.com/edgarSucre/flight/token"
)

const (
	authHeader   = "Authorization"
	bearerPrefix = "Bearer"
)

func jwtMiddleware(next http.Handler, tokenMaker token.Maker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/asset") ||
			r.URL.Path == "/favicon.ico" ||
			r.URL.Path == "/" ||
			r.URL.Path == "/health" ||
			r.URL.Path == "/user/login" {

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

		username, err := tokenMaker.VerifyToken(auth[len(bearerPrefix)+1:])
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), flight.UserCtxKey, username)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
