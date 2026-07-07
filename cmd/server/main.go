package main

import (
	"log"
	"net/http"

	"houseauto/internal/api"
	"houseauto/internal/config"
	"houseauto/internal/database"
	"houseauto/internal/mqttclient"
	"houseauto/internal/sse"
)

func main() {
	// 1. Cargar Configuración
	cfg := config.Load()

	// 2. Inicializar Base de Datos SQLite
	db := database.Init()

	// 3. Inicializar el Broker de SSE para retransmitir logs
	sseBroker := sse.NewBroker()

	// 4. Inicializar y conectar Cliente MQTT (le pasamos db para guardar eventos)
	mqttClient := mqttclient.NewClient(cfg, sseBroker, db)

	// 5. Registrar Rutas HTTP (le pasamos db para el historial)
	api.RegisterHandlers(cfg, mqttClient, sseBroker, db)

	// 6. Iniciar Servidor Web
	log.Println("[+] Servidor Go inicializado en puerto 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error de red en servidor HTTP: %v", err)
	}
}
