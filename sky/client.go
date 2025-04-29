package sky

import (
	"github.com/edgarSucre/flight"
)

type Client struct {
	baseUrl      string
	rapidAPIKey  string
	rapidApiHost string
	requester    flight.Requester
}

func NewClient(rapidApiKey, rapidApiHost, baseUrl string, r flight.Requester) *Client {
	return &Client{
		baseUrl:      baseUrl,
		rapidApiHost: rapidApiHost,
		rapidAPIKey:  rapidApiKey,
		requester:    r,
	}
}
