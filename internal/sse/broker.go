package sse

import (
	"fmt"
	"net/http"
	"sync"
)

type Broker struct {
	clients    map[chan string]bool
	clientsMu  sync.Mutex
	lastStatus string
	statusMu   sync.RWMutex
}

func NewBroker() *Broker {
	return &Broker{
		clients:    make(map[chan string]bool),
		lastStatus: "offline", // estado por defecto hasta recibir señal
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

// UpdateStatus guarda el estado actual y lo envía a los clientes
func (b *Broker) UpdateStatus(status string) {
	b.statusMu.Lock()
	b.lastStatus = status
	b.statusMu.Unlock()

	msg := fmt.Sprintf(`{"type":"status","value":"%s"}`, status)
	b.Broadcast(msg)
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

	// Enviar el último estado conocido inmediatamente al nuevo cliente sin generar log extra
	b.statusMu.RLock()
	currentStatus := b.lastStatus
	b.statusMu.RUnlock()
	fmt.Fprintf(w, "data: {\"type\":\"init_status\",\"value\":\"%s\"}\n\n", currentStatus)

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
