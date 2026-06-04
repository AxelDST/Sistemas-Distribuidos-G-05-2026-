package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sd-iot/pkg/coap"
	"sd-iot/pkg/mqtt"
	"sd-iot/pkg/nodo"
	"sd-iot/pkg/sensor"
)

func main() {
	log.Println("Iniciando Nodo Smart Campus...")

	// Cargar configuración del entorno
	config := nodo.CargarConfiguracion()
	log.Printf("Configuración: %+v\n", config)

	// Inicializar simulador de sensor
	sim := sensor.NuevoSimulador()

	// TODO 1: Crear y conectar el cliente MQTT con Testamento (LWT).
	// El testamento debe publicar en nodo/{id}/estado el mensaje {"estado":"offline"} con QoS 1 y retained=true.
	// Además, publicar un mensaje retenido {"estado":"online"} tras conectar.
	clienteMQTT, err := mqtt.NuevoCliente(config)
	if err != nil {
		log.Fatalf("Error creando cliente MQTT: %v", err)
	}
	if err := clienteMQTT.Conectar(); err != nil {
		log.Fatalf("Error conectando MQTT: %v", err)
	}
	log.Println("Cliente MQTT conectado")

	// TODO 2: Iniciar loop de lecturas periódicas.
	// Cada config.IntervaloSegundos segundos:
	//   - Obtener lectura del simulador.
	//   - Publicar en campus/{edificio}/{aula}/sensor/temperatura con QoS 1.
	go clienteMQTT.PublicarLecturas(sim, config)

	// TODO 3: Suscribirse a comandos del actuador.
	// Tópico: campus/{edificio}/{aula}/actuador/cmd
	// Al recibir un comando, imprimirlo por consola y ejecutar la acción (ej: encender alarma).
	if err := clienteMQTT.SuscribirComandos(config); err != nil {
		log.Fatalf("Error suscribiendo a comandos: %v", err)
	}

	// TODO 4: Iniciar servidor CoAP en puerto UDP 5683.
	// Recursos obligatorios:
	//   GET /temperatura  -> devuelve la última lectura del sensor en JSON.
	//   PUT /config       -> actualiza configuración local (ej: {"modo":"manual"}).
	//   GET /config       -> devuelve la configuración actual en JSON.
	servidorCoAP := coap.NuevoServidor(sim, config)
	go servidorCoAP.Iniciar()
	log.Println("Servidor CoAP iniciado en :5683")

	// Esperar señal de terminación
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	fmt.Println("\nCerrando nodo limpiamente...")
	clienteMQTT.Desconectar()
	log.Println("Nodo finalizado")
}
