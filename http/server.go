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
	"github.com/rs/cors"
)

func NewServer(
	providers []flight.Provider,
	tokenMaker token.Maker,
	config util.Config,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, providers, tokenMaker, config)

	var handler http.Handler = mux

	c := cors.New(cors.Options{
		AllowedOrigins:   config.AllowedOrigins,
		AllowCredentials: true,
		// Debug:            true,
		// Logger:           log.Default(),
		AllowedHeaders: []string{"authorization", "content-type"},
	})

	handler = jwtMiddleware(handler, tokenMaker)
	handler = c.Handler(handler)
	// handler = logHandler(handler)

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
