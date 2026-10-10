package sse

import (
	"log"
	"sync"
)

type SSEHub struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{}
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[chan []byte]struct{}),
	}
}

func (hub *SSEHub) Register(client chan []byte) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	hub.clients[client] = struct{}{}
}

func (hub *SSEHub) Unregister(client chan []byte) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	delete(hub.clients, client)
}

func (hub *SSEHub) Broadcast(data []byte) {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	for client := range hub.clients {
		select {
		case client <- data:
		default:
			log.Println("SSE client is too slow, event skipped")
		}
	}
}
