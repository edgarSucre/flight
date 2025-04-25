package flight

import (
	"context"
	"errors"
	"time"
)

type (
	Info struct {
		Duration time.Duration
		Price    float64
		Agent    string
	}

	InfoByPrice []Info

	Cabin string
)

func (s InfoByPrice) Len() int {
	return len(s)
}

func (s InfoByPrice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s InfoByPrice) Less(i, j int) bool {
	return s[i].Price < s[j].Price
}

func (s InfoByPrice) Cheapest() Info {
	var cheapest Info

	if len(s) == 0 {
		return cheapest
	}

	cheapest = s[0]

	for _, v := range s {
		if v.Price < cheapest.Price {
			cheapest = v
		}
	}

	return cheapest
}

func (s InfoByPrice) Fastest() Info {
	var fastest Info

	if len(s) == 0 {
		return fastest
	}

	fastest = s[0]

	for _, v := range s {
		if v.Duration < fastest.Duration {
			fastest = v
		}
	}

	return fastest
}

const (
	Business       Cabin = "Business"
	Economy        Cabin = "Economy"
	First          Cabin = "First"
	PremiumEconomy Cabin = "Premium_Economy"
)

var ValidCabins = map[string]struct{}{
	"Business":        {},
	"Economy":         {},
	"First":           {},
	"Premium_Economy": {},
}

type SearchParams struct {
	ArrivalAirport   string
	CabinClass       string
	Currency         string
	DepartureAirport string
	DepartureDate    time.Time
	NumAdults        int
	NumChildren      int
	NumInfants       int
}

var (
	ErrInvalidArrivalCode   = errors.New("destination airport code must be a valid 3 letter IATA code")
	ErrInvalidDepartureCode = errors.New("origin airport code must be a valid 3 letter IATA code")
	ErrInvalidDepartureDate = errors.New("departure date is invalid")
)

func (params SearchParams) Validate() error {
	if len(params.ArrivalAirport) < 3 || len(params.ArrivalAirport) > 3 {
		return ErrInvalidArrivalCode
	}

	if len(params.DepartureAirport) < 3 || len(params.DepartureAirport) > 3 {
		return ErrInvalidDepartureCode
	}

	today := time.Now()
	minDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)

	if params.DepartureDate.Before(minDate) {
		return ErrInvalidDepartureDate
	}

	return nil
}

func SetDefaultParams(params SearchParams) SearchParams {
	if params.NumAdults == 0 {
		params.NumAdults = 1
	}

	if len(params.Currency) == 0 {
		params.Currency = "USD"
	}

	if len(params.CabinClass) == 0 {
		params.CabinClass = "Economy"
	}

	return params
}

type Provider interface {
	Search(ctx context.Context, params SearchParams) ([]Info, error)
}
