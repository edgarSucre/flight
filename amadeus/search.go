package amadeus

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
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
	AccessToken string
	ExpiresAt   time.Time
}

func (t tokenEntry) valid() bool {
	return t.ExpiresAt.After(time.Now())
}

func (c *Client) Search(ctx context.Context, params flight.SearchParams) ([]flight.Info, error) {
	params = flight.SetDefaultParams(params)

	username := ctx.Value(flight.UserCtxKey).(string)

	token, err := c.getToken(ctx, username)
	if err != nil {
		return nil, err
	}

	c.tokens[username] = token

	r, err := c.search(ctx, params, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("amadeus.search: %w", err)
	}

	return r.buildResponse(), nil
}

const (
	flightEndpoint = "/v2/shopping/flight-offers"
	oauthEndpoint  = "v1/security/oauth2/token"
)

func (c *Client) getToken(ctx context.Context, username string) (tokenEntry, error) {
	if token, ok := c.tokens[username]; ok && token.valid() {
		return token, nil
	}

	var payload bytes.Buffer
	payload.WriteString(
		fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s", c.key, c.secret),
	)

	url := fmt.Sprintf("%s/%s", c.baseUrl, oauthEndpoint)

	reader, _, err := c.requester.MakeRequest(
		ctx,
		http.MethodPost,
		url,
		&payload,
		util.URLEncoded,
	)

	if err != nil {
		return tokenEntry{}, fmt.Errorf("could not authenticate with amadeus: %w", err)
	}

	var auth authResponse

	err = util.JsonDecode(reader, &auth)
	if err != nil {
		return tokenEntry{}, fmt.Errorf("could not authenticate with amadeus: %w", err)
	}

	return tokenEntry{
		AccessToken: auth.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(auth.ExpiresIn) * time.Second),
	}, nil
}

type (
	segment struct {
		CarrierCode string `json:"carrierCode"`
	}

	itinerary struct {
		Duration string    `json:"duration"`
		Segments []segment `json:"segments"`
	}
)

func (it itinerary) duration() time.Duration {
	d := it.Duration
	if strings.HasPrefix(d, "PT") {
		d = strings.TrimPrefix(d, "PT")
	}

	d = strings.ToLower(d)
	t, _ := time.ParseDuration(d)

	return t
}

func (it itinerary) buildInfo(agentsIdx map[string]string) flight.Info {
	var b strings.Builder

	for _, v := range it.Segments {
		if agent, ok := agentsIdx[v.CarrierCode]; ok {
			b.WriteString(fmt.Sprintf("%s-", agent))
		}
	}

	return flight.Info{
		Agent:    strings.TrimSuffix(b.String(), "-"),
		Duration: it.duration(),
	}
}

type (
	price struct {
		Total string `json:"total"`
	}

	data struct {
		Itineraries []itinerary `json:"itineraries"`
		Price       price       `json:"price"`
	}
)

type (
	dictionary struct {
		Carriers map[string]string `json:"carriers"`
	}

	searchResponse struct {
		Data         []data     `json:"data"`
		Dictionaries dictionary `json:"dictionaries"`
	}
)

func (resp searchResponse) buildResponse() []flight.Info {
	if len(resp.Data) == 0 {
		return nil
	}

	agentsIdx := make(map[string]string)

	for k, v := range resp.Dictionaries.Carriers {
		agentsIdx[k] = v
	}

	infoResponse := make([]flight.Info, 0, len(resp.Data))

	for _, d := range resp.Data {

		price, err := strconv.ParseFloat(d.Price.Total, 64)
		if err != nil {
			continue
		}

		for _, it := range d.Itineraries {
			info := it.buildInfo(agentsIdx)
			info.Price = price

			infoResponse = append(infoResponse, info)
		}
	}

	return infoResponse

}

func (c *Client) search(
	ctx context.Context,
	params flight.SearchParams,
	token string,
) (searchResponse, error) {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("%s/%s", c.baseUrl, flightEndpoint))
	builder.WriteString(fmt.Sprintf("?originLocationCode=%s", params.DepartureAirport))
	builder.WriteString(fmt.Sprintf("&destinationLocationCode=%s", params.ArrivalAirport))
	builder.WriteString(fmt.Sprintf("&departureDate=%s", params.DepartureDate.Format(time.DateOnly)))
	builder.WriteString(fmt.Sprintf("&adults=%v", params.NumAdults))
	builder.WriteString(fmt.Sprintf("&travelClass=%s", "ECONOMY"))
	builder.WriteString(fmt.Sprintf("&currencyCode=%s", params.Currency))

	url := builder.String()

	reader, _, err := c.requester.MakeRequest(
		ctx,
		http.MethodGet,
		url,
		nil,
		map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)},
	)

	if err != nil {
		return searchResponse{}, fmt.Errorf("request to amadeus API failed: %w", err)
	}

	var payload searchResponse

	return payload, util.JsonDecode(reader, &payload)
}
