package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/edgarSucre/flight"
)

var ErrInvalidDate = errors.New("invalid date, must be in YYYY-MM-DD format")

const (
	queryParamDeparture     = "origin"
	queryParamArrival       = "destination"
	queryParamDepartureDate = "date"
)

func handleSearch(providers []flight.Provider) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params, err := buildParams(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data, err := lookUpFlights(r.Context(), providers, params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := buildResponse(data)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		encoder := json.NewEncoder(w)
		if err := encoder.Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func buildParams(r *http.Request) (flight.SearchParams, error) {
	queryParams := r.URL.Query()

	departureDate, err := time.Parse(time.DateOnly, queryParams.Get(queryParamDepartureDate))
	if err != nil {
		return flight.SearchParams{}, ErrInvalidDate
	}

	params := flight.SearchParams{
		ArrivalAirport:   queryParams.Get(queryParamArrival),
		DepartureAirport: queryParams.Get(queryParamDeparture),
		DepartureDate:    departureDate,
	}

	return params, params.Validate()
}
