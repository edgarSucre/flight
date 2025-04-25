package http

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJWTMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte{})
	})

	createErr, verifyErr := make(chan error, 2), make(chan error, 2)

	tokenMaker := tokenMakerMock(t, createErr, verifyErr)

	tests := []struct {
		name  string
		req   *http.Request
		stub  func()
		check func(t *testing.T, response *http.Response)
	}{
		{
			"noAuth",
			newChallengeRequest(t, "/nowhere", "", true),
			nil,
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusUnauthorized)

				payload, _ := io.ReadAll(response.Body)

				errMsg := "missing Authorization header\n"
				assert.Equal(t, errMsg, string(payload))
			},
		},
		{
			"noBearerPrefix",
			newChallengeRequest(t, "/nowhere", "", false),
			nil,
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusUnauthorized)

				payload, _ := io.ReadAll(response.Body)

				errMsg := "missing Bearer token\n"
				assert.Equal(t, errMsg, string(payload))
			},
		},
		{
			"invalidToken",
			newChallengeRequest(t, "/nowhere", "token", false),
			func() {
				verifyErr <- fmt.Errorf("invalid token")
			},
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusUnauthorized)

				payload, _ := io.ReadAll(response.Body)

				errMsg := "invalid token\n"
				assert.Equal(t, errMsg, string(payload))
			},
		},
		{
			"validToken",
			newChallengeRequest(t, "/nowhere", "token", false),
			nil,
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusNoContent)

				payload, _ := io.ReadAll(response.Body)

				assert.Empty(t, payload)
			},
		},
		{
			"skipped",
			newChallengeRequest(t, "/health", "token", false),
			nil,
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusNoContent)

				payload, _ := io.ReadAll(response.Body)

				assert.Empty(t, payload)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.stub != nil {
				tt.stub()
			}

			rec := httptest.NewRecorder()

			handler := jwtMiddleware(handler, tokenMaker)

			handler.ServeHTTP(rec, tt.req)

			tt.check(t, rec.Result())
		})
	}
}

func newChallengeRequest(t *testing.T, url, token string, skipHeader bool) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, url, nil)

	if skipHeader {
		return req
	}

	if len(token) == 0 {
		req.Header.Add(authHeader, "token")

		return req
	}

	req.Header.Add(authHeader, fmt.Sprintf("%s %s", bearerPrefix, token))

	return req
}
