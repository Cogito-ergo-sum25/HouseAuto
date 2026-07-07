package api

import (
	"fmt"
	"net/http"

	"houseauto/internal/config"
	"houseauto/internal/mqttclient"
	"houseauto/internal/sse"
)

func RegisterHandlers(cfg *config.Config, mqttClient *mqttclient.Client, sseBroker *sse.Broker) {
	// Servir archivos estáticos del frontend
	fs := http.FileServer(http.Dir("./frontend/public"))
	http.Handle("/", fs)

	// Endpoint Server-Sent Events (SSE) para logs en tiempo real
	http.HandleFunc("/api/logs", sseBroker.Handler)

	// Endpoint API para accionar el portón
	http.HandleFunc("/api/abrir", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		password := r.FormValue("password")
		if password != cfg.AuthPassword {
			http.Error(w, "Contraseña incorrecta", http.StatusUnauthorized)
			return
		}

		if err := mqttClient.PublishAbrir(); err != nil {
			http.Error(w, "Error al comunicar con el portón", http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, "¡Señal de apertura enviada con éxito al broker MQTT!")
	})
}
