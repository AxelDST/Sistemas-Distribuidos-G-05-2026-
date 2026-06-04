package mqtt

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"sd-iot/pkg/nodo"
	"sd-iot/pkg/sensor"
)

// Cliente encapsula la conexión MQTT del nodo.
type Cliente struct {
	config   nodo.Configuracion
	interno  mqtt.Client
	opciones *mqtt.ClientOptions
}

// TODO 1: NuevoCliente crea la configuración inicial del cliente MQTT.
// Debe:
//   1a. Construir el tópico del testamento: nodo/{id}/estado
//   1b. Configurar el mensaje del testamento como {"estado":"offline"} con QoS 1 y retained=true.
//   1c. Configurar ClientID único, timeout de conexión y reconexión automática.
//
// Sugerencia: usar mqtt.NewClientOptions().AddBroker(...).SetClientID(...).SetWill(...)
func NuevoCliente(config nodo.Configuracion) (*Cliente, error) {
	// COMPLETAR
	return nil, fmt.Errorf("TODO: implementar NuevoCliente")
}

// TODO 2: Conectar establece la sesión con el broker.
// Tras conectar, debe publicar un mensaje retenido {"estado":"online"} en nodo/{id}/estado.
func (c *Cliente) Conectar() error {
	// COMPLETAR
	return fmt.Errorf("TODO: implementar Conectar")
}

// TODO 3: PublicarLecturas envía periódicamente las lecturas del sensor.
// Debe:
//   3a. Construir el tópico: campus/{edificio}/{aula}/sensor/temperatura
//   3b. En un ticker cada config.IntervaloSegundos, llamar sim.Leer(), serializar a JSON y publicar con QoS 1.
//   El JSON debe tener: {"nodo_id": ..., "temperatura": ..., "unidad":"C", "timestamp":"..."}
func (c *Cliente) PublicarLecturas(sim *sensor.Simulador, config nodo.Configuracion) {
	// COMPLETAR
}

// TODO 4: SuscribirComandos se une al tópico de actuadores y procesa mensajes.
// Debe:
//   4a. Suscribirse al tópico campus/{edificio}/{aula}/actuador/cmd con QoS 1.
//   4b. En el callback, deserializar el JSON, imprimir el comando recibido y simular la ejecución.
//   Ejemplo de payload esperado: {"accion":"encender_alarma", "origen":"dashboard"}
func (c *Cliente) SuscribirComandos(config nodo.Configuracion) error {
	// COMPLETAR
	return fmt.Errorf("TODO: implementar SuscribirComandos")
}

// TODO 5: Desconectar cierra limpiamente la sesión MQTT.
// Sugerencia: publicar estado offline retenido antes de desconectar.
func (c *Cliente) Desconectar() {
	// COMPLETAR
}
