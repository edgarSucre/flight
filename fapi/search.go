package fapi

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/edgarSucre/flight"
	"github.com/edgarSucre/flight/util"
)

var (
	ErrContextCancelled = errors.New("request time out")
)

func (c *Client) Search(ctx context.Context, params flight.SearchParams) ([]flight.Info, error) {
	params = flight.SetDefaultParams(params)

	r, err := c.search(ctx, params)

	if err != nil {
		return nil, fmt.Errorf("fapi.search: %w", err)
	}

	return r.buildResponse(), nil
}

type (
	price struct {
		Amount float64 `json:"amount"`
	}

	pricingOption struct {
		AgentIDs []string `json:"agent_ids"`
		Price    price    `json:"price"`
	}

	leg struct {
		Duration int    `json:"duration"`
		ID       string `json:"id"`
	}

	agent struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
)

type itinerary struct {
	ID             string          `json:"id"`
	LegIDs         []string        `json:"leg_ids"`
	PricingOptions []pricingOption `json:"pricing_options"`
}

func (it itinerary) cheapest() pricingOption {
	cheapest := pricingOption{
		Price: price{
			Amount: math.MaxFloat64,
		},
	}

	for _, opt := range it.PricingOptions {
		if opt.Price.Amount < cheapest.Price.Amount {
			cheapest = opt
		}
	}

	return cheapest
}

func (it itinerary) buildInfo(
	legsIdx map[string]int,
	agentsIdx map[string]string,
) flight.Info {
	duration := time.Duration(legsIdx[it.LegIDs[0]])

	cheapest := it.cheapest()

	info := flight.Info{
		Price:    cheapest.Price.Amount,
		Agent:    agentsIdx[cheapest.AgentIDs[0]],
		Duration: time.Duration(time.Minute * duration),
	}

	return info
}

type searchResponse struct {
	Itineraries []itinerary `json:"itineraries"`
	Legs        []leg       `json:"legs"`
	Agents      []agent     `json:"agents"`
}

func (resp searchResponse) buildResponse() []flight.Info {
	if len(resp.Itineraries) == 0 {
		return nil
	}

	legsIdx := make(map[string]int)
	agentsIdx := make(map[string]string)

	for _, l := range resp.Legs {
		legsIdx[l.ID] = l.Duration
	}

	for _, a := range resp.Agents {
		agentsIdx[a.ID] = a.Name
	}

	infoResponse := make([]flight.Info, len(resp.Itineraries))

	for i, it := range resp.Itineraries {
		infoResponse[i] = it.buildInfo(legsIdx, agentsIdx)
	}

	return infoResponse
}

func (c *Client) search(ctx context.Context, params flight.SearchParams) (searchResponse, error) {
	url := fmt.Sprintf(
		"%s/%s/%s/%s/%s/%v/%v/%v/%s/%s",
		c.host,
		c.key,
		params.DepartureAirport,
		params.ArrivalAirport,
		params.DepartureDate.Format(time.DateOnly),
		params.NumAdults,
		params.NumChildren,
		params.NumInfants,
		"Economy",
		params.Currency,
	)

	reader, _, err := c.requester.MakeRequest(
		ctx,
		http.MethodGet,
		url,
		nil,
		util.JSON,
	)

	if err != nil {
		return searchResponse{}, fmt.Errorf("request to flight API failed: %w", err)
	}

	var payload searchResponse

	return payload, util.JsonDecode(reader, &payload)
}
