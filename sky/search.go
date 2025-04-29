package sky

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/edgarSucre/flight"
	"github.com/edgarSucre/flight/util"
)

func (c *Client) Search(ctx context.Context, params flight.SearchParams) ([]flight.Info, error) {
	params = flight.SetDefaultParams(params)

	r, err := c.search(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("sky-scanner.search: %w", err)
	}

	return r.buildResponse(), nil
}

type (
	price struct {
		Raw float64 `json:"raw"`
	}

	marketing struct {
		Name string `json:"name"`
	}

	carrier struct {
		Marketing []marketing `json:"marketing"`
	}

	leg struct {
		Carriers          carrier `json:"carriers"`
		DurationInMinutes int     `json:"durationInMinutes"`
	}
)

func (ca carrier) provider() string {
	var b strings.Builder

	for _, v := range ca.Marketing {
		b.WriteString(fmt.Sprintf("%s-", v.Name))
	}

	return strings.TrimSuffix(b.String(), "-")
}

func (l leg) duration() time.Duration {
	return time.Duration(l.DurationInMinutes) * time.Minute
}

func (l leg) provider() string {
	return l.Carriers.provider()
}

type (
	itinerary struct {
		Price price `json:"price"`
		Legs  []leg `json:"legs"`
	}
)

func (it itinerary) buildInfo() flight.Info {
	leg := it.Legs[0]

	return flight.Info{
		Agent:    leg.provider(),
		Duration: leg.duration(),
		Price:    it.Price.Raw,
	}
}

type (
	data struct {
		Itineraries []itinerary `json:"itineraries"`
	}

	searchResponse struct {
		Data data `json:"data"`
	}
)

func (resp searchResponse) buildResponse() []flight.Info {
	if len(resp.Data.Itineraries) == 0 {
		return nil
	}

	infoResponse := make([]flight.Info, len(resp.Data.Itineraries))

	for i, it := range resp.Data.Itineraries {
		infoResponse[i] = it.buildInfo()
	}

	return infoResponse

}

const (
	flightEndpoint     = "/flights/search-one-way"
	rapidAPIHostHeader = "x-rapidapi-host"
	rapidAPIKeyHeader  = "x-rapidapi-key"
)

func (c *Client) search(
	ctx context.Context,
	params flight.SearchParams,
) (searchResponse, error) {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("%s/%s", c.baseUrl, flightEndpoint))
	builder.WriteString(fmt.Sprintf("?fromEntityId=%s", params.DepartureAirport))
	builder.WriteString(fmt.Sprintf("&toEntityId=%s", params.ArrivalAirport))
	builder.WriteString(fmt.Sprintf("&departDate=%s", params.DepartureDate.Format(time.DateOnly)))
	builder.WriteString(fmt.Sprintf("&adults=%v", params.NumAdults))
	builder.WriteString(fmt.Sprintf("&cabinClass=%s", "economy"))
	builder.WriteString(fmt.Sprintf("&currency=%s", params.Currency))

	url := builder.String()

	reader, _, err := c.requester.MakeRequest(
		ctx,
		http.MethodGet,
		url,
		nil,
		map[string]string{
			rapidAPIHostHeader: c.rapidApiHost,
			rapidAPIKeyHeader:  c.rapidAPIKey,
		},
	)

	if err != nil {
		return searchResponse{}, fmt.Errorf("request to sky scanner API failed: %w", err)
	}

	var payload searchResponse

	return payload, util.JsonDecode(reader, &payload)
}
