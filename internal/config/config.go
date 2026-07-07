package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MQTTBroker   string
	AuthPassword string
}

func Load() *Config {
	// Intentar cargar .env para desarrollo local (si no existe, usa variables inyectadas por Docker)
	_ = godotenv.Load()

	mqttBroker := os.Getenv("MQTT_BROKER")
	authPassword := os.Getenv("PORTON_PASSWORD")

	if mqttBroker == "" || authPassword == "" {
		log.Fatal("[-] Error: Faltan variables de entorno críticas (MQTT_BROKER o PORTON_PASSWORD)")
	}

	return &Config{
		MQTTBroker:   mqttBroker,
		AuthPassword: authPassword,
	}
}
