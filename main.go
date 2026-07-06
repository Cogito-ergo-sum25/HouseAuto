package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	// Leer variables de entorno inyectadas por Docker
	mqttBroker := os.Getenv("MQTT_BROKER")
	authPassword := os.Getenv("PORTON_PASSWORD")

	if mqttBroker == "" || authPassword == "" {
		log.Fatal("[-] Error: Faltan variables de entorno críticas (MQTT_BROKER o PORTON_PASSWORD)")
	}

	// Servir la interfaz web
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.Execute(w, nil)
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

		// Conectar al cliente Mosquitto interno
		opts := mqtt.NewClientOptions().AddBroker(mqttBroker).SetClientID("porton_web_backend")
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Printf("Error MQTT local: %v", token.Error())
			http.Error(w, "Error de comunicación interna", http.StatusInternalServerError)
			return
		}
		defer client.Disconnect(250)

		// Publicar el token de apertura
		token := client.Publish("casa/porton/abrir", 0, false, "1")
		token.Wait()

		fmt.Fprint(w, "¡Señal de apertura enviada con éxito!")
	})

	log.Println("[+] Servidor Go inicializado en puerto 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error de red en servidor HTTP: %v", err)
	}
}