package http

import (
	"net/http"

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
		AllowedHeaders:   []string{"authorization", "content-type"},
	})

	handler = jwtMiddleware(handler, tokenMaker)
	handler = c.Handler(handler)

	return handler
}
