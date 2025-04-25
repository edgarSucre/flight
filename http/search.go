package http

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"sync"

	"github.com/edgarSucre/flight"
)

func handleSearch(providers []flight.Provider) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: set params
		params := flight.SearchParams{}

		data, err := lookUpFlightsInfo(r.Context(), providers, params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := buildResponse(data)

		w.WriteHeader(http.StatusOK)

		encoder := json.NewEncoder(w)
		if err := encoder.Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

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

// TODO: add context to provider calls

func lookUpFlightsInfo(
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

			info, err := p.Search(params)

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
		// TODO: return error here
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
		comparison[i] = buildInfoResponse(v)
	}

	return lookUpResponse{
		Cheapest:   buildInfoResponse(cheapest),
		Fastest:    buildInfoResponse(fastest),
		Comparison: comparison,
	}
}

func buildInfoResponse(nfo flight.Info) flightInfo {
	return flightInfo{
		Agent:    nfo.Agent,
		Duration: nfo.Duration.String(),
		Price:    nfo.Price,
	}
}
