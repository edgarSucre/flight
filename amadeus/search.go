package amadeus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/edgarSucre/flight"
	"github.com/edgarSucre/flight/util"
)

type authResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type tokenEntry struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   time.Time
}

func (c *Client) Search(ctx context.Context, params flight.SearchParams) ([]flight.Info, error) {
	params = flight.SetDefaultParams(params)

	// TODO: move ctx to domain, so you can use here
	username := ctx.Value("username").(string)
	if payload, ok := c.tokens[username]; ok {
		var token tokenEntry

		err := json.Unmarshal(payload, &token)
		if err != nil {
			// get new token
		}

		if time.Now().After(token.ExpiresAt) {
			// get new token
		}

		// use same token
	}

	return nil, nil
}

const (
	flightEndpoint = "/v2/shopping/flight-offers"
)

func (c *Client) search(
	ctx context.Context,
	params flight.SearchParams,
	token string,
) ([]string, error) {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("%s/%s", c.baseUrl, flightEndpoint))
	builder.WriteString(fmt.Sprintf("?originLocationCode=%s", params.DepartureAirport))
	builder.WriteString(fmt.Sprintf("&destinationLocationCode=%s", params.ArrivalAirport))
	builder.WriteString(fmt.Sprintf("&departureDate=%s", params.DepartureDate))
	builder.WriteString(fmt.Sprintf("&adults=%s", params.NumAdults))

	// TODO: make sure cabin class match
	builder.WriteString(fmt.Sprintf("&travelClass=%s", params.CabinClass))
	builder.WriteString(fmt.Sprintf("&currencyCode=%s", params.Currency))

	url := builder.String()

	reader, _, err := c.requester.MakeRequest(
		ctx,
		http.MethodGet,
		url,
		nil,
		map[string]string{"accept": "application/vnd.amadeus+json"},
	)

	if err != nil {
		return searchResponse{}, fmt.Errorf("request to flight API failed: %w", err)
	}

	var payload searchResponse

	return payload, util.JsonDecode(reader, &payload)
}
