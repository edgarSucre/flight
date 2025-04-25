package http

import (
	"context"
	"errors"
	"sort"
	"sync"

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
	errCh := make(chan bool)
	done := make(chan bool)

	data := make([]flight.Info, 0)

	go func() {
		for {
			select {
			case <-errCh:
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

	for _, p := range providers {
		go func() {
			defer wg.Done()

			info, err := p.Search(ctx, params)

			if err != nil {
				errCh <- true
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
	fastest := byPrice.Fastest()

	comparison := make([]flightInfo, len(data))

	for i, v := range byPrice {
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
