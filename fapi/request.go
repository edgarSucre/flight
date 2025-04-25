package fapi

import (
	"bytes"
	"context"
	"io"
	"os"
	"path"
)

// temporary file for testing

type Requester struct{}

func (Requester) MakeRequest(
	ctx context.Context,
	method string,
	url string,
	body io.Reader,
	headers map[string]string,
) (io.Reader, int, error) {
	fullPath, _ := os.Getwd()

	if os.Getenv("debug") == "true" {
		fullPath = path.Join(fullPath, "..")
	}

	fullPath = path.Join(fullPath, "fapi_mock_response.json")

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, 0, err
	}

	return bytes.NewReader(data), 0, nil
}
