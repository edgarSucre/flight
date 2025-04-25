package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/edgarSucre/flight"
)

const (
	queryParamDeparture     = "origin"
	queryParamArrival       = "destination"
	queryParamDepartureDate = "date"
)

func handleSearch(providers []flight.Provider) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()

		departureDate, err := time.Parse(time.DateOnly, queryParams.Get(queryParamDepartureDate))
		if err != nil {
			http.Error(w, "invalid date, must be in YYYY-MM-DD format", http.StatusBadRequest)
			return
		}

		params := flight.SearchParams{
			ArrivalAirport:   queryParams.Get(queryParamArrival),
			DepartureAirport: queryParams.Get(queryParamDeparture),
			DepartureDate:    departureDate,
		}

		if err := params.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data, err := lookUpFlights(r.Context(), providers, params)
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
