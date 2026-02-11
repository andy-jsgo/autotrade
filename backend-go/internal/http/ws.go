package http

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

func ServeWS(addr string) {
	h := func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for t := range ticker.C {
			msg := map[string]any{"type": "heartbeat", "at": t.UTC().Format(time.RFC3339)}
			if err := c.WriteJSON(msg); err != nil {
				return
			}
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", h)
	log.Printf("ws listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("ws stopped: %v", err)
	}
}
