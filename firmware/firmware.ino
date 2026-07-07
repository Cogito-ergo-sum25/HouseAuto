#include <WiFi.h>
#include <PubSubClient.h>
#include <RCSwitch.h>
#include "secrets.h"

// En lugar de escribir el texto plano, le asignas las macros definidas en secrets.h
const char* ssid = SECRET_SSID;
const char* password = SECRET_PASS;

// Tu subdominio o IP de internet para el Broker MQTT
const char* mqtt_server = "192.168.1.170";
const int mqtt_port = 1883; 

// ... (Todo el resto de tu código de inicialización, setup y loops se queda exactamente igual)
WiFiClient espClient;
PubSubClient client(espClient);
RCSwitch mySwitch = RCSwitch();

void logMessage(String msg) {
  Serial.println(msg);
  if (client.connected()) {
    client.publish("casa/porton/logs", msg.c_str());
  }
}

void setup() {
  Serial.begin(115200);
  delay(1000); 
  
  setup_wifi();
  
  client.setServer(mqtt_server, mqtt_port);
  client.setCallback(callback); 
  
  // Transmisor FS1000A al GPIO 13
  mySwitch.enableTransmit(13);
  mySwitch.setProtocol(6);
  mySwitch.setPulseLength(477);
  mySwitch.setRepeatTransmit(15);
  
  Serial.println("\n[-] Esperando mensajes desde broker.cereva.lat...");
  // Nota: aquí aún no estamos conectados a MQTT, así que no usamos logMessage
}

void setup_wifi() {
  delay(10);
  Serial.print("Conectando Wi-Fi...");
  WiFi.begin(ssid, password);

  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("\n[+] Wi-Fi Conectado exitosamente.");
}

void callback(char* topic, byte* payload, unsigned int length) {
  String messageTemp;
  for (int i = 0; i < length; i++) {
    messageTemp += (char)payload[i];
  }

  if (String(topic) == "casa/porton/abrir" && messageTemp == "1") {
    logMessage("[+] Comando de apertura recibido desde internet!");
    mySwitch.send(108628005, 28); // Tu código clonado de 28 bits
    logMessage("[+] Señal RF 433 MHz enviada al motor SEG.");
  }
}

void reconnect() {
  while (!client.connected()) {
    Serial.print("Conectando al Mosquitto Local...");
    
    // ID único para el actuador, LWT (Last Will and Testament)
    if (client.connect("ESP32C3_Porton_Actuator", "casa/porton/status", 1, true, "offline")) {
      logMessage("¡Conectado a MQTT con éxito!");
      // Publicar estado online retenido
      client.publish("casa/porton/status", "online", true);
      client.subscribe("casa/porton/abrir");
    } else {
      Serial.print(" falló, rc=");
      Serial.print(client.state());
      Serial.println(" Reintentando en 5 segundos...");
      delay(5000);
    }
  }
}

void loop() {
  if (!client.connected()) {
    reconnect();
  }
  client.loop();
}