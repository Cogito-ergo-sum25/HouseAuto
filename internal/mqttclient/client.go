package mqttclient

import (
	"log"

	"houseauto/internal/config"
	"houseauto/internal/sse"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	conn mqtt.Client
}

func NewClient(cfg *config.Config, sseBroker *sse.Broker) *Client {
	opts := mqtt.NewClientOptions().AddBroker(cfg.MQTTBroker).SetClientID("porton_web_backend_persistent")

	opts.OnConnect = func(c mqtt.Client) {
		log.Println("[+] Conectado al broker MQTT. Suscribiendo a logs...")
		if token := c.Subscribe("casa/porton/logs", 0, func(c mqtt.Client, m mqtt.Message) {
			sseBroker.Broadcast(string(m.Payload()))
		}); token.Wait() && token.Error() != nil {
			log.Printf("[-] Error al suscribirse a logs: %v", token.Error())
		}
	}

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("[-] Error inicial MQTT local: %v", token.Error())
	}

	return &Client{
		conn: mqttClient,
	}
}

func (c *Client) PublishAbrir() error {
	token := c.conn.Publish("casa/porton/abrir", 0, false, "1")
	token.Wait()
	return token.Error()
}
