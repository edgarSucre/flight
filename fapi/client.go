package fapi

type Client struct {
	host string
	key  string
	env  string
}

// TODO: remove environment from client
func NewClient(apiKey, apiHost, environment string) *Client {
	return &Client{
		env:  environment,
		host: apiHost,
		key:  apiKey,
	}
}
