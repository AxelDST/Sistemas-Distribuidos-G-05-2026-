package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"sd-iot/pkg/nodo"
	"sd-iot/pkg/sensor"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Cliente encapsula la conexión MQTT del nodo.
type Cliente struct {
	config   nodo.Configuracion
	interno  mqtt.Client
	opciones *mqtt.ClientOptions
}

// TODO 1: NuevoCliente crea la configuración inicial del cliente MQTT.
// Debe:
//
//	1a. Construir el tópico del testamento: nodo/{id}/estado
//	1b. Configurar el mensaje del testamento como {"estado":"offline"} con QoS 1 y retained=true.
//	1c. Configurar ClientID único, timeout de conexión y reconexión automática.
//
// Sugerencia: usar mqtt.NewClientOptions().AddBroker(...).SetClientID(...).SetWill(...)
func NuevoCliente(config nodo.Configuracion) (*Cliente, error) {
	topicoEstado := fmt.Sprintf("nodo/%s/estado", config.ID)
	broker := config.BrokerMQTT
	if !strings.Contains(broker, "://") {
		broker = "tcp://" + broker
	}
	clientID := fmt.Sprintf("%s-%d", config.ID, time.Now().UnixNano())
	opciones := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetWill(topicoEstado, `{"estado":"offline"}`, 1, true).
		SetConnectTimeout(5 * time.Second).
		SetAutoReconnect(true).
		SetCleanSession(true)

	cliente := mqtt.NewClient(opciones)
	return &Cliente{
		config:   config,
		interno:  cliente,
		opciones: opciones,
	}, nil
}

// TODO 2: Conectar establece la sesión con el broker.
// Tras conectar, debe publicar un mensaje retenido {"estado":"online"} en nodo/{id}/estado.
func (c *Cliente) Conectar() error {
	if token := c.interno.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return c.interno.Publish(
		fmt.Sprintf("nodo/%s/estado", c.config.ID),
		1,
		true,
		`{"estado":"online"}`,
	).Error()
}

// TODO 3: PublicarLecturas envía periódicamente las lecturas del sensor.
// Debe:
//
//	3a. Construir el tópico: campus/{edificio}/{aula}/sensor/temperatura
//	3b. En un ticker cada config.IntervaloSegundos, llamar sim.Leer(), serializar a JSON y publicar con QoS 1.
//	El JSON debe tener: {"nodo_id": ..., "temperatura": ..., "unidad":"C", "timestamp":"..."}
func (c *Cliente) PublicarLecturas(sim *sensor.Simulador, config nodo.Configuracion) {
	topico := fmt.Sprintf("campus/%s/%s/sensor/temperatura", config.Edificio, config.Aula)
	ticker := time.NewTicker(config.IntervaloSegundos)
	defer ticker.Stop()

	for range ticker.C {
		lectura := sim.Leer()
		payload := struct {
			NodoID      string  `json:"nodo_id"`
			Temperatura float64 `json:"temperatura"`
			Unidad      string  `json:"unidad"`
			Timestamp   string  `json:"timestamp"`
		}{
			NodoID:      config.ID,
			Temperatura: lectura,
			Unidad:      "C",
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}
		data, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Error serializando lectura: %v", err)
			continue
		}

		if token := c.interno.Publish(topico, 1, false, data); token.Wait() && token.Error() != nil {
			log.Printf("Error publicando lectura: %v", token.Error())
		}
	}
}

// TODO 4: SuscribirComandos se une al tópico de actuadores y procesa mensajes.
// Debe:
//
//	4a. Suscribirse al tópico campus/{edificio}/{aula}/actuador/cmd con QoS 1.
//	4b. En el callback, deserializar el JSON, imprimir el comando recibido y simular la ejecución.
//	Ejemplo de payload esperado: {"accion":"encender_alarma", "origen":"dashboard"}
func (c *Cliente) SuscribirComandos(config nodo.Configuracion) error {
	topico := fmt.Sprintf("campus/%s/%s/actuador/cmd", config.Edificio, config.Aula)
	callback := func(_ mqtt.Client, msg mqtt.Message) {
		var comando struct {
			Accion string `json:"accion"`
			Origen string `json:"origen"`
		}
		if err := json.Unmarshal(msg.Payload(), &comando); err != nil {
			log.Printf("Comando invalido (%s): %s", err, string(msg.Payload()))
			return
		}
		log.Printf("Comando recibido: accion=%s origen=%s", comando.Accion, comando.Origen)
		if comando.Accion != "" {
			log.Printf("Ejecutando accion: %s", comando.Accion)
		}
	}

	if token := c.interno.Subscribe(topico, 1, callback); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// TODO 5: Desconectar cierra limpiamente la sesión MQTT.
// Sugerencia: publicar estado offline retenido antes de desconectar.
func (c *Cliente) Desconectar() {
	if c == nil || c.interno == nil {
		return
	}
	_ = c.interno.Publish(
		fmt.Sprintf("nodo/%s/estado", c.config.ID),
		1,
		true,
		`{"estado":"offline"}`,
	).Error()
	c.interno.Disconnect(250)
}
