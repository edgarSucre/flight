package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var (
	JSON = map[string]string{"Content-Type": "application/json"}
)

type HttpRequester struct{}

func (HttpRequester) MakeRequest(
	ctx context.Context,
	method string,
	url string,
	body io.Reader,
	headers map[string]string,
) (io.Reader, int, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		url,
		body,
	)

	if err != nil {
		return nil, 0, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if resp != nil {
			return nil, resp.StatusCode, err
		}

		return nil, 0, err
	}

	return resp.Body, resp.StatusCode, nil
}

func JsonDecode[T any](resp io.Reader, v T) error {
	if err := json.NewDecoder(resp).Decode(&v); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}

	return nil
}
