package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/edgarSucre/flight/token"
	"github.com/edgarSucre/flight/util"
)

type (
	loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	loginResponse struct {
		AccessToken string `json:"access_token"`
	}
)

// TODO: change default duration
const duration = time.Duration(time.Hour * 200)

func handleLogin(tokenMaker token.Maker, config util.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest

		decoder := json.NewDecoder(r.Body)

		if err := decoder.Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if config.DefaultUserName != req.Username || config.DefaultUserPassword != req.Password {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		token, err := tokenMaker.CreateToken(req.Username, duration)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := loginResponse{
			AccessToken: token,
		}

		encoder := json.NewEncoder(w)
		if err := encoder.Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
