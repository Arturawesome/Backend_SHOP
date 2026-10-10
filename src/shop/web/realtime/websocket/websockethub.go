package websockethub

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

func (socket *WebSocketHub) Register(conn *websocket.Conn) {
	socket.mu.Lock()
	defer socket.mu.Unlock()
	socket.clients[conn] = struct{}{}
}

func (socket *WebSocketHub) Unregister(conn *websocket.Conn) {
	socket.mu.Lock()
	defer socket.mu.Unlock()
	delete(socket.clients, conn)
}

func (socket *WebSocketHub) Broadcast(data []byte) {
	socket.mu.RLock()
	defer socket.mu.RUnlock()

	for conn := range socket.clients {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("websocket write:", err)
		}
	}
}
