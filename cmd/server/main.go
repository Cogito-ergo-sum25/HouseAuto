package main

import (
	"log"
	"net/http"

	"houseauto/internal/api"
	"houseauto/internal/config"
	"houseauto/internal/mqttclient"
	"houseauto/internal/sse"
)

func main() {
	// 1. Cargar Configuración
	cfg := config.Load()

	// 2. Inicializar el Broker de SSE para retransmitir logs
	sseBroker := sse.NewBroker()

	// 3. Inicializar y conectar Cliente MQTT
	mqttClient := mqttclient.NewClient(cfg, sseBroker)

	// 4. Registrar Rutas HTTP
	api.RegisterHandlers(cfg, mqttClient, sseBroker)

	// 5. Iniciar Servidor Web
	log.Println("[+] Servidor Go inicializado en puerto 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error de red en servidor HTTP: %v", err)
	}
}
