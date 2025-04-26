package fapi

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/edgarSucre/flight/util"
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
	data, err := os.ReadFile(util.FilePath("fapi_mock_response.json"))
	if err != nil {
		return nil, 0, err
	}

	return bytes.NewReader(data), 0, nil
}
