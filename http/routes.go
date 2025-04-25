package http

import (
	"net/http"

	"github.com/edgarSucre/flight/token"

	"github.com/edgarSucre/flight"
	"github.com/edgarSucre/flight/util"
)

func addRoutes(
	mux *http.ServeMux,
	providers []flight.Provider,
	tokenMaker token.Maker,
	config util.Config,
) {
	mux.Handle("GET /flights/search", handleSearch(providers))
	mux.Handle("POST /user/login", handleLogin(tokenMaker, config))
	mux.HandleFunc("GET /health", handleHealth)
	mux.Handle("/", http.NotFoundHandler())
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	w.Write(nil)
}
