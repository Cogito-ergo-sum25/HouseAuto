package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	clients   = make(map[chan string]bool)
	clientsMu sync.Mutex
)

// broadcastLog envía un mensaje a todos los clientes web conectados por SSE
func broadcastLog(msg string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for clientChan := range clients {
		// Non-blocking send para no bloquear si el cliente es lento
		select {
		case clientChan <- msg:
		default:
		}
	}
}

func main() {
	// Leer variables de entorno inyectadas por Docker
	mqttBroker := os.Getenv("MQTT_BROKER")
	authPassword := os.Getenv("PORTON_PASSWORD")

	if mqttBroker == "" || authPassword == "" {
		log.Fatal("[-] Error: Faltan variables de entorno críticas (MQTT_BROKER o PORTON_PASSWORD)")
	}

	// Configurar conexión MQTT persistente
	opts := mqtt.NewClientOptions().AddBroker(mqttBroker).SetClientID("porton_web_backend_persistent")
	
	opts.OnConnect = func(c mqtt.Client) {
		log.Println("[+] Conectado al broker MQTT. Suscribiendo a logs...")
		if token := c.Subscribe("casa/porton/logs", 0, func(c mqtt.Client, m mqtt.Message) {
			// Cuando llega un log desde el ESP32, retransmitirlo a la web
			broadcastLog(string(m.Payload()))
		}); token.Wait() && token.Error() != nil {
			log.Printf("[-] Error al suscribirse a logs: %v", token.Error())
		}
	}
	
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("[-] Error inicial MQTT local: %v", token.Error())
	}

	// Servir la interfaz web principal
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.Execute(w, nil)
	})

	// Endpoint Server-Sent Events (SSE) para logs en tiempo real
	http.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming no soportado", http.StatusInternalServerError)
			return
		}

		messageChan := make(chan string, 10)
		
		clientsMu.Lock()
		clients[messageChan] = true
		clientsMu.Unlock()

		// Limpiar al desconectar el cliente
		defer func() {
			clientsMu.Lock()
			delete(clients, messageChan)
			clientsMu.Unlock()
			close(messageChan)
		}()

		// Mensaje de bienvenida al conectar
		fmt.Fprintf(w, "data: [Web] Conectado a la consola. Esperando eventos...\n\n")
		flusher.Flush()

		// Bucle de escucha
		for {
			select {
			case msg := <-messageChan:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			case <-r.Context().Done():
				return // Cliente cerró la conexión
			}
		}
	})

	// Endpoint API para accionar el portón
	http.HandleFunc("/api/abrir", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		// Validar contraseña del formulario contra el .env
		password := r.FormValue("password")
		if password != authPassword {
			http.Error(w, "Contraseña incorrecta", http.StatusUnauthorized)
			return
		}

		// Publicar el token de apertura usando la conexión persistente
		token := mqttClient.Publish("casa/porton/abrir", 0, false, "1")
		token.Wait()

		fmt.Fprint(w, "¡Señal de apertura enviada con éxito al broker MQTT!")
	})

	log.Println("[+] Servidor Go inicializado en puerto 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error de red en servidor HTTP: %v", err)
	}
}