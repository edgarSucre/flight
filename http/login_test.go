package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/edgarSucre/flight/token"
	"github.com/edgarSucre/flight/util"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	createErr, verifyErr := make(chan error, 2), make(chan error, 2)

	tokenMaker := tokenMakerMock(t, createErr, verifyErr)

	config := util.Config{
		DefaultUserName:     "test",
		DefaultUserPassword: "pass",
	}

	tests := []struct {
		name  string
		req   *http.Request
		stub  func()
		check func(t *testing.T, response *http.Response)
	}{
		{
			"decodeError",
			newLoginRequest(t, "", ""),
			nil,
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusBadRequest)

				payload, _ := io.ReadAll(response.Body)

				errMsg := "bad credentials\n"
				assert.Equal(t, errMsg, string(payload))
			},
		},
		{
			"invalidCredentials",
			newLoginRequest(t, "test", ""),
			nil,
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusUnauthorized)

				payload, _ := io.ReadAll(response.Body)

				errMsg := "invalid credentials\n"
				assert.Equal(t, errMsg, string(payload))
			},
		},
		{
			"tokenErr",
			newLoginRequest(t, "test", "pass"),
			func() {
				createErr <- fmt.Errorf("fail now")
			},
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusInternalServerError)

				payload, _ := io.ReadAll(response.Body)

				errMsg := "could not login, server error\n"
				assert.Equal(t, errMsg, string(payload))
			},
		},
		{
			"success",
			newLoginRequest(t, "test", "pass"),
			nil,
			func(t *testing.T, response *http.Response) {
				assert.Equal(t, response.StatusCode, http.StatusOK)

				decoder := json.NewDecoder(response.Body)

				var payload loginResponse

				err := decoder.Decode(&payload)
				assert.NoError(t, err)

				assert.Equal(t, "signed token", payload.AccessToken)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.stub != nil {
				tt.stub()
			}

			rec := httptest.NewRecorder()

			handler := handleLogin(tokenMaker, config)

			handler.ServeHTTP(rec, tt.req)

			tt.check(t, rec.Result())
		})
	}
}

type mockMaker struct {
	createErr chan error
	verifyErr chan error
}

func (maker mockMaker) CreateToken(username string, duration time.Duration) (string, error) {
	select {
	case err := <-maker.createErr:
		return "", err
	default:
		return "signed token", nil
	}
}

func (maker mockMaker) VerifyToken(token string) (string, error) {
	select {
	case err := <-maker.verifyErr:
		return "", err
	default:
		return "username", nil
	}
}

func tokenMakerMock(t *testing.T, createErr, verifyErr chan error) token.Maker {
	t.Helper()

	return mockMaker{createErr, verifyErr}
}

func newLoginRequest(t *testing.T, username string, password string) *http.Request {
	t.Helper()

	return httptest.NewRequest(
		http.MethodPost,
		"/user/login",
		bytes.NewReader(newBody(t, username, password)),
	)
}

func newBody(t *testing.T, username, password string) []byte {
	t.Helper()

	if len(username) == 0 && len(password) == 0 {
		return []byte{}
	}

	creds := loginRequest{
		Username: username,
		Password: password,
	}

	payload, _ := json.Marshal(creds)

	return payload
}
