package http

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"

	"github.com/edgarSucre/flight"
	"github.com/edgarSucre/flight/token"
	"github.com/edgarSucre/flight/util"
)

func NewServer(
	providers []flight.Provider,
	tokenMaker token.Maker,
	config util.Config,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, providers, tokenMaker, config)

	var handler http.Handler = mux

	handler = jwtMiddleware(handler, tokenMaker)

	return handler
}

func logHandler(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		x, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		log.Println(fmt.Sprintf("%q", x))
		rec := httptest.NewRecorder()

		fn.ServeHTTP(rec, r)

		log.Println(fmt.Sprintf("%q", rec.Body))
	})
}
