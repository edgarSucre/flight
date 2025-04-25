package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/edgarSucre/flight"
	"github.com/edgarSucre/flight/fapi"
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
		fapi.NewClient(config.FlightAPIKey, config.FlightAPIURL, config.Environment),
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
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
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
