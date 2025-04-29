package http

import (
	"log"
	"net/http"
	"time"

	"github.com/edgarSucre/flight"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func handleWS(providers []flight.Provider) http.Handler {
	conns := make(map[string]*websocket.Conn)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := ctx.Value(flight.UserCtxKey).(string)

		if conn, ok := conns[user]; ok {
			conn.Close()
		}

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade fail", err)

			http.Error(w, "upgrade fail", http.StatusBadRequest)
			return
		}

		conns[user] = c

		defer c.Close()

		ticker := time.Tick(time.Second * 30)

		for {
			<-ticker

			params, err := buildParams(r)
			if err != nil {
				msg := "could not read request params"
				log.Println(msg, err)

				c.WriteMessage(1, []byte(msg))

				break
			}

			data, err := lookUpFlights(r.Context(), providers, params)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			resp := buildResponse(data)
			if err := c.WriteJSON(resp); err != nil {

			}
		}
	})
}
