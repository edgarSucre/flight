package http

import (
	"net/http"

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
