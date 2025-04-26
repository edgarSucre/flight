package amadeus

import (
	"github.com/edgarSucre/flight"
)

type Client struct {
	baseUrl   string
	key       string
	requester flight.Requester
	secret    string
	tokens    map[string]tokenEntry
}

func NewClient(apiKey, apiSecret, baseUrl string, r flight.Requester) *Client {
	return &Client{
		baseUrl:   baseUrl,
		key:       apiKey,
		requester: r,
		secret:    apiSecret,
		tokens:    make(map[string]tokenEntry),
	}
}
