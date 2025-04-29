package http

import (
	"context"
	"errors"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/edgarSucre/flight"
)

var (
	ErrCouldNotFetchFlights = errors.New("unable to look up flight prices")
)

type (
	flightInfo struct {
		Agent    string  `json:"agent"`
		Duration string  `json:"duration"`
		Price    float64 `json:"price"`
	}

	lookUpResponse struct {
		Cheapest   flightInfo   `json:"cheapest"`
		Comparison []flightInfo `json:"comparison"`
		Fastest    flightInfo   `json:"fastest"`
	}
)

func lookUpFlights(
	ctx context.Context,
	providers []flight.Provider,
	params flight.SearchParams,
) ([]flight.Info, error) {
	n := len(providers)
	workingProviders := n

	infoCh := make(chan []flight.Info, n)
	errCh := make(chan error)
	done := make(chan bool)

	data := make([]flight.Info, 0)

	go func() {
		for {
			select {
			case err := <-errCh:
				log.Println(err)
				workingProviders--
			case info := <-infoCh:
				data = append(data, info...)
			case <-done:
				return
			}
		}
	}()

	var wg sync.WaitGroup

	wg.Add(len(providers))

	for i, p := range providers {
		go func() {
			defer wg.Done()

			t := time.Now()

			log.Printf("sending request for provider # %v", i)

			info, err := p.Search(ctx, params)

			log.Printf("request for provider # %v took %s", i, time.Since(t))

			if err != nil {
				errCh <- err
				return
			}

			infoCh <- info
		}()
	}

	wg.Wait()
	done <- true

	if workingProviders <= 0 {
		return nil, ErrCouldNotFetchFlights
	}

	return data, nil
}

func buildResponse(data []flight.Info) lookUpResponse {
	if len(data) == 0 {
		return lookUpResponse{}
	}

	byPrice := flight.InfoByPrice(data)
	sort.Sort(byPrice)

	cheapest := byPrice[0]

	fastest := flight.Info{
		Duration: time.Duration(math.MaxInt),
	}

	comparison := make([]flightInfo, len(data))

	for i, v := range byPrice {
		if v.Duration < fastest.Duration {
			fastest = v
		} else if v.Duration == fastest.Duration && v.Price < fastest.Price {
			fastest = v
		}

		comparison[i] = buildInfo(v)
	}

	return lookUpResponse{
		Cheapest:   buildInfo(cheapest),
		Fastest:    buildInfo(fastest),
		Comparison: comparison,
	}
}

func buildInfo(nfo flight.Info) flightInfo {
	return flightInfo{
		Agent:    nfo.Agent,
		Duration: nfo.Duration.String(),
		Price:    nfo.Price,
	}
}
