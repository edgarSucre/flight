package fapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/edgarSucre/flight"
)

var (
	ErrContextCancelled = errors.New("request time out")
)

func (c *Client) Search(ctx context.Context, params flight.SearchParams) ([]flight.Info, error) {
	if c.env == "dev" {
		return c.fakeIT()
	}

	params = flight.SetDefaultParams(params)

	searchResponse, err := search(
		ctx,
		c.host,
		c.key,
		params,
	)

	if err != nil {
		return nil, fmt.Errorf("fapi.search: %w", err)
	}

	return searchResponse.buildResponse(), nil
}

type (
	price struct {
		Amount float64 `json:"amount"`
	}

	pricingOptions struct {
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
	ID             string           `json:"id"`
	LegIDs         []string         `json:"leg_ids"`
	PricingOptions []pricingOptions `json:"pricing_options"`
}

func (it itinerary) cheapest() pricingOptions {
	cheapest := it.PricingOptions[0]

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

func search(ctx context.Context, host, key string, params flight.SearchParams) (searchResponse, error) {
	url := fmt.Sprintf(
		"%s/%s/%s/%s/%s/%v/%v/%v/%s/%s",
		host,
		key,
		params.DepartureAirport,
		params.ArrivalAirport,
		params.DepartureDate,
		params.NumAdults,
		params.NumChildren,
		params.NumInfants,
		params.CabinClass,
		params.Currency,
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return searchResponse{}, fmt.Errorf("http.NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Duration(time.Second * 15),
	}

	respCh := make(chan *http.Response)
	errCh := make(chan error)

	go func() {
		resp, err := client.Do(req)
		if err != nil {
			errCh <- fmt.Errorf("http.Client.Do: %w", err)
		}

		respCh <- resp
	}()

	select {
	case resp := <-respCh:
		return decode(resp)
	case err = <-errCh:
		return searchResponse{}, err
	case <-ctx.Done():
		return searchResponse{}, ErrContextCancelled
	}
}

func decode(response *http.Response) (searchResponse, error) {
	decoder := json.NewDecoder(response.Body)

	var payload searchResponse

	if err := decoder.Decode(&payload); err != nil {
		return searchResponse{}, fmt.Errorf("json.Decode: %w", err)
	}

	return payload, nil
}

func (c *Client) fakeIT() ([]flight.Info, error) {
	data, err := os.ReadFile("/home/edgar/Documents/code/me/flight/response_jfk_sfo.json")
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(data))

	var jsonResponse searchResponse

	if err := decoder.Decode(&jsonResponse); err != nil {
		return nil, err
	}

	return jsonResponse.buildResponse(), nil
}
