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
	"github.com/edgarSucre/flight/amadeus"
	"github.com/edgarSucre/flight/fapi"
	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	flightApiErr := make(chan error, 2)
	flightApiProvider := flightAPI(t, flightApiErr)

	amadeusApiErr := make(chan error, 2)
	amadeusApiProvider := amadeusAPI(t, amadeusApiErr)

	validCtx := context.Background()
	validCtx = context.WithValue(validCtx, flight.UserCtxKey, "username")
	expiredCtx, _ := context.WithTimeout(validCtx, time.Nanosecond)

	tests := []struct {
		name  string
		req   *http.Request
		stub  func()
		check func(t *testing.T, response *http.Response)
	}{
		{
			"invalidDate",
			newSerchRequest(t, validCtx, "JFK", "SFO", "2030-13-02"),
			nil,
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
			nil,
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
			nil,
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
			nil,
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
			func() {
				// make amadeus fail so cheaper comes from flight api
				amadeusApiErr <- fmt.Errorf("amadeus failed")
			},
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
					Price:    98.97,
				}

				assert.Equal(t, cheapest, payload.Cheapest)
				assert.Equal(t, fastest, payload.Fastest)
			},
		},
		{
			"successWithAmadeusApi",
			newSerchRequest(t, validCtx, "JFK", "SFO", "2030-10-02"),
			func() {
				// make flight api fail so cheaper comes from amadeus
				flightApiErr <- fmt.Errorf("flight api failed")
			},
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusOK)

				decoder := json.NewDecoder(response.Body)

				var payload lookUpResponse

				err := decoder.Decode(&payload)
				assert.NoError(t, err)

				cheapest := flightInfo{
					Agent:    "ALASKA AIRLINES",
					Duration: "6h28m0s",
					Price:    149.25,
				}

				fastest := flightInfo{
					Agent:    "JETBLUE AIRWAYS",
					Duration: "6h17m0s",
					Price:    173.45,
				}

				assert.Equal(t, cheapest, payload.Cheapest)
				assert.Equal(t, fastest, payload.Fastest)
			},
		},
		{
			"comparison",
			newSerchRequest(t, validCtx, "JFK", "SFO", "2030-10-02"),
			nil, // all providers will submit their results
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
					Price:    98.97,
				}

				assert.Equal(t, cheapest, payload.Cheapest)
				assert.Equal(t, fastest, payload.Fastest)
			},
		},
		{
			"allFail",
			newSerchRequest(t, validCtx, "JFK", "SFO", "2030-10-02"),
			func() {
				// make all fail
				flightApiErr <- fmt.Errorf("flight api failed")
				amadeusApiErr <- fmt.Errorf("amadeus api failed")
			},
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusInternalServerError)

				payload, _ := io.ReadAll(response.Body)
				errMsg := "unable to look up flight prices\n"
				assert.Equal(t, errMsg, string(payload))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			handler := handleSearch([]flight.Provider{flightApiProvider, amadeusApiProvider})

			if tt.stub != nil {
				tt.stub()
			}

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
		path: "fapi_mock_response.json",
	}

	return fapi.NewClient("key", "host", r)
}

func amadeusAPI(
	t *testing.T,
	err chan error,
) flight.Provider {
	t.Helper()

	r := requester{
		err:  err,
		path: "amadeus_mock_response.json",
	}

	return amadeus.NewClient("key", "secret", "url", r)
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
		return loadTestData(r.path)
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
