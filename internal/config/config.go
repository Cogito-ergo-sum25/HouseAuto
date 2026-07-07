package config

import (
	"log"
	"os"
)

type Config struct {
	MQTTBroker   string
	AuthPassword string
}

func Load() *Config {
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
