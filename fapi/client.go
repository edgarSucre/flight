package fapi

import "github.com/edgarSucre/flight"

type Client struct {
	host      string
	key       string
	requester flight.Requester
}

func NewClient(apiKey, apiHost string, r flight.Requester) *Client {
	return &Client{
		host:      apiHost,
		key:       apiKey,
		requester: r,
	}
}
