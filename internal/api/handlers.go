package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"houseauto/internal/config"
	"houseauto/internal/database"
	"houseauto/internal/mqttclient"
	"houseauto/internal/sse"
)

func RegisterHandlers(cfg *config.Config, mqttClient *mqttclient.Client, sseBroker *sse.Broker, db *database.DB) {
	// Generar un hash único de versión cada vez que el servidor arranca
	version := fmt.Sprintf("%d", time.Now().Unix())

	fs := http.FileServer(http.Dir("./frontend/public"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		
		// Si es HTML, inyectar el cache buster dinámicamente
		if strings.HasSuffix(path, ".html") {
			fullPath := filepath.Join("frontend", "public", path)
			tmpl, err := template.ParseFiles(fullPath)
			if err != nil {
				// Dejar que fs maneje el 404
				fs.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			tmpl.Execute(w, map[string]string{"Version": version})
			return
		}

		// Servir CSS, JS e imágenes estáticamente
		fs.ServeHTTP(w, r)
	})

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

		db.SaveEvent("log", "Comando de apertura enviado desde interfaz web")
		fmt.Fprint(w, "¡Señal de apertura enviada con éxito al broker MQTT!")
	})

	// Endpoint para obtener el historial de eventos (consola)
	http.HandleFunc("/api/history", func(w http.ResponseWriter, r *http.Request) {
		events, err := db.GetRecentEvents(50)
		if err != nil {
			http.Error(w, "Error al obtener historial", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	})
}
