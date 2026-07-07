package sse

import (
	"fmt"
	"net/http"
	"sync"
)

type Broker struct {
	clients   map[chan string]bool
	clientsMu sync.Mutex
}

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[chan string]bool),
	}
}

// Broadcast envía un mensaje a todos los clientes web conectados
func (b *Broker) Broadcast(msg string) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()
	for clientChan := range b.clients {
		select {
		case clientChan <- msg:
		default:
		}
	}
}

// Handler para registrar y servir a los clientes SSE
func (b *Broker) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming no soportado", http.StatusInternalServerError)
		return
	}

	messageChan := make(chan string, 10)

	b.clientsMu.Lock()
	b.clients[messageChan] = true
	b.clientsMu.Unlock()

	defer func() {
		b.clientsMu.Lock()
		delete(b.clients, messageChan)
		b.clientsMu.Unlock()
		close(messageChan)
	}()

	fmt.Fprintf(w, "data: {\"type\":\"log\",\"message\":\"[Web] Conectado a la consola. Esperando eventos...\"}\n\n")
	flusher.Flush()

	for {
		select {
		case msg := <-messageChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
