package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/edgarSucre/flight"
	"github.com/edgarSucre/flight/fapi"
	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	errCh := make(chan error)

	flightApiProvider := flightAPI(t, errCh)

	validCtx := context.Background()
	expiredCtx, _ := context.WithTimeout(validCtx, time.Nanosecond)

	tests := []struct {
		name  string
		req   *http.Request
		check func(t *testing.T, response *http.Response)
	}{
		{
			"invalidDate",
			newSerchRequest(t, validCtx, "JFK", "SFO", "2030-13-02"),
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusBadRequest)

				payload, _ := io.ReadAll(response.Body)
				errMsg := "invalid date, must be in YYYY-MM-DD format\n"
				assert.Equal(t, errMsg, string(payload))
			},
		},
		{
			"invalidOrigin",
			newSerchRequest(t, validCtx, "SFSL", "SFO", "2030-10-02"),
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusBadRequest)

				payload, _ := io.ReadAll(response.Body)
				errMsg := fmt.Sprintf("%s\n", flight.ErrInvalidDepartureCode)
				assert.Equal(t, errMsg, string(payload))
			},
		},
		{
			"invalidDestination",
			newSerchRequest(t, validCtx, "JFK", "SFOS", "2030-10-02"),
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusBadRequest)

				payload, _ := io.ReadAll(response.Body)
				errMsg := fmt.Sprintf("%s\n", flight.ErrInvalidArrivalCode)
				assert.Equal(t, errMsg, string(payload))
			},
		},
		{
			"timeOut",
			newSerchRequest(t, expiredCtx, "JFK", "SFO", "2030-10-02"),
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusInternalServerError)

				payload, _ := io.ReadAll(response.Body)
				errMsg := fmt.Sprintf("%s\n", ErrCouldNotFetchFlights)
				assert.Equal(t, errMsg, string(payload))
			},
		},
		{
			"successWithFlightApi",
			newSerchRequest(t, validCtx, "JFK", "SFO", "2030-10-02"),
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusOK)

				decoder := json.NewDecoder(response.Body)

				var payload lookUpResponse

				err := decoder.Decode(&payload)
				assert.NoError(t, err)

				cheapest := flightInfo{
					Agent:    "BudgetAir",
					Duration: "6h20m0s",
					Price:    98.96,
				}

				fastest := flightInfo{
					Agent:    "BudgetAir",
					Duration: "6h10m0s",
					Price:    98.96,
				}

				assert.Equal(t, payload.Cheapest, cheapest)
				assert.Equal(t, payload.Fastest, fastest)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			handler := handleSearch([]flight.Provider{flightApiProvider})

			handler.ServeHTTP(rec, tt.req)

			tt.check(t, rec.Result())
		})
	}
}

func flightAPI(
	t *testing.T,
	err chan error,
) flight.Provider {
	t.Helper()

	r := requester{
		err:  err,
		path: "",
	}

	return fapi.NewClient("key", "host", r)
}

type requester struct {
	err  chan error
	path string
}

func (r requester) MakeRequest(
	ctx context.Context,
	method string,
	url string,
	body io.Reader,
	headers map[string]string,
) (io.Reader, int, error) {
	select {
	case err := <-r.err:
		return nil, 0, err
	case <-ctx.Done():
		return nil, 0, fmt.Errorf("time out")
	default:
		return loadTestData("fapi_mock_response.json")
	}
}

func loadTestData(fileName string) (io.Reader, int, error) {
	fullPath, _ := os.Getwd()

	fullPath = path.Join(fullPath, "..", fileName)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, 0, err
	}

	return bytes.NewReader(data), 0, nil
}

func newSerchRequest(
	t *testing.T,
	ctx context.Context,
	orig string,
	dest string,
	date string,
) *http.Request {
	url := fmt.Sprintf("/flights/search?origin=%s&destination=%s&date=%s", orig, dest, date)

	return httptest.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}
