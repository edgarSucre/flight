package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/edgarSucre/flight"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Date        string `json:"date"`
}

func handleWS(providers []flight.Provider) http.Handler {
	// conns := make(map[string]*websocket.Conn)
	var ws *websocket.Conn

	close := make(chan struct{})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ws != nil {
			ws.Close()
			close <- struct{}{}
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade fail", err)

			http.Error(w, "upgrade fail", http.StatusBadRequest)
			return
		}

		defer ws.Close()

		req := make(chan flight.SearchParams)

		ticker := time.NewTicker(time.Second * 30)

		go func() {
			var params *flight.SearchParams

			for {
				select {
				case p := <-req:
					params = &p
				case <-ticker.C:
					if params != nil {
						data, err := lookUpFlights(r.Context(), providers, *params)
						if err != nil {
							log.Println("could not fetch flights", err)
							continue
						}

						response := buildResponse(data)

						payload, err := json.Marshal(response)
						if err != nil {
							log.Println("could not encode flights", err)
							continue
						}

						if err := ws.WriteMessage(1, payload); err != nil {
							log.Println("could send websocket message", err)
						}
					}
				case <-close:
					log.Println("closing ws..")
					return
				}
			}
		}()

		for {
			messageType, message, err := ws.ReadMessage()
			if err != nil {
				log.Println("could not read from the socket:", err)

				closeError := new(websocket.CloseError)

				if errors.As(err, &closeError) {
					close <- struct{}{}
					return
				}

				continue
			}

			if messageType == 8 {
				close <- struct{}{}
				break
			}

			if messageType == 9 || string(message) == "ping" {
				ws.WriteMessage(10, []byte("pong"))
			}

			var payload wsRequest

			if err := json.Unmarshal(message, &payload); err != nil {
				log.Println("could not parse from the socket:", err)
			}

			departureDate, err := time.Parse(time.DateOnly, payload.Date)
			if err != nil {
				log.Println("could not parse from the socket:", err)
			}

			req <- flight.SearchParams{
				ArrivalAirport:   payload.Destination,
				Currency:         "USD",
				DepartureAirport: payload.Origin,
				DepartureDate:    departureDate,
				NumAdults:        1,
				NumChildren:      0,
				NumInfants:       0,
			}
		}
	})
}
