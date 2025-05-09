package http

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/edgarSucre/flight/token"
	"github.com/edgarSucre/flight/util"
	"golang.org/x/crypto/bcrypt"
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

const duration = time.Duration(time.Minute * 15)

func handleLogin(tokenMaker token.Maker, config util.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest

		decoder := json.NewDecoder(r.Body)

		if err := decoder.Decode(&req); err != nil {
			http.Error(w, "bad credentials", http.StatusBadRequest)
			return
		}

		err := bcrypt.CompareHashAndPassword([]byte(config.DefaultUserPassword), []byte(req.Password))

		if config.DefaultUserName != req.Username || err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		token, err := tokenMaker.CreateToken(req.Username, duration)
		if err != nil {
			log.Println(err)

			http.Error(w, "could not login, server error", http.StatusInternalServerError)
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
