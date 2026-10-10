package web

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func (app *Application) sseHandler(w http.ResponseWriter, r *http.Request) {
	client := make(chan []byte, 10)
	app.sseHub.Register(client)
	defer app.sseHub.Unregister(client)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// for immediate info sending
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(
			w,
			"streaming unsupported",
			http.StatusInternalServerError,
		)
		return
	}

	for {
		select {
		case data := <-client:
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}

}

func (app *Application) webSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
	if err != nil {
		app.logger.Error("WebSocket upgrade failed", "error", err)
		return
	}
	app.wsHub.Register(conn)
	defer func() {
		app.wsHub.Unregister(conn)
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

	}

}
