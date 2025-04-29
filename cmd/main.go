package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/edgarSucre/flight"
	"github.com/edgarSucre/flight/amadeus"
	"github.com/edgarSucre/flight/fapi"
	"github.com/edgarSucre/flight/sky"

	fHttp "github.com/edgarSucre/flight/http"
	"github.com/edgarSucre/flight/token"
	"github.com/edgarSucre/flight/util"
)

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	config, err := util.LoadConfig(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading configuration: %s\n", err)
		return err
	}

	providers := []flight.Provider{
		amadeus.NewClient(
			config.AmadeusAPIKey,
			config.AmadeusAPISecret,
			config.AmadeusAPIBaseURL,
			util.HttpRequester{},
		),
		fapi.NewClient(config.FlightAPIKey,
			config.FlightAPIURL,
			util.HttpRequester{},
		),
		sky.NewClient(
			config.AmadeusAPIKey,
			config.SkyScannerRapidAPIHost,
			config.SkyScannerRapidAPIBaseURL,
			util.HttpRequester{},
		),
	}

	tokenMaker, err := token.NewJWTMaker(config.JwtKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating JWT Maker: %s\n", err)
		return err
	}

	srv := fHttp.NewServer(providers, tokenMaker, config)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Host, config.Port),
		Handler: srv,
	}

	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		cert := util.FilePath("certs/server-cert.pem")
		key := util.FilePath("certs/server-key.pem")

		if err := httpServer.ListenAndServeTLS(cert, key); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
			cancel()

			return
		}

		fmt.Fprintln(os.Stdout, " server shutting down..")
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()

	wg.Wait()
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
