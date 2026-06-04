package main

import (
	"fmt"
	"math/rand"
	"net/rpc/jsonrpc"
	"os"
	"time"

	// TODO: importar "time" para el loop de lecturas y heartbeats
	// TODO: importar "net/rpc/jsonrpc" para conectar con JSON
	// TODO: importar "sd-comunicacion/pkg/telemetria"
	// TODO: importar "sd-comunicacion/pkg/protocolo"
	// TODO: importar "sd-comunicacion/internal/detector"

	"sd-comunicacion/internal/detector"
	"sd-comunicacion/pkg/protocolo"
	"sd-comunicacion/pkg/telemetria"
)

// TODO 14: Implementar una funcion LlamarRPC.
// Conecta al servidor via jsonrpc.Dial, llama el metodo, cierra la conexion y retorna el error.
// Firma sugerida:
//   func LlamarRPC(addr, metodo string, args, resp interface{}) error
// Necesitaras: import "net/rpc/jsonrpc"

func LlamarRPC(addr, metodo string, args, resp interface{}) error {
	conn, err := jsonrpc.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.Call(metodo, args, resp)
}

func main() {
	// Configuracion
	servidorAddr := os.Getenv("SERVIDOR")
	if servidorAddr == "" {
		servidorAddr = "localhost:1234"
	}
	nombre := os.Getenv("NOMBRE")
	if nombre == "" {
		nombre = "cliente-1"
	}
	heartbeatPuerto := os.Getenv("HEARTBEAT_PUERTO")
	if heartbeatPuerto == "" {
		heartbeatPuerto = ":4002"
	}

	fmt.Printf("[CLIENTE %s] Iniciando...\n", nombre)

	// TODO 15: Iniciar el receptor de heartbeats para monitorear al servidor.
	// Crear un detector.NuevoReceptor(heartbeatPuerto, 6*time.Second) y
	// ejecutar Escuchar() en una goroutine.
	// Necesitaras: import "sd-comunicacion/internal/detector" y "time"
	receptor := detector.NuevoReceptor(heartbeatPuerto, 6*time.Second)
	go receptor.Escuchar()

	// TODO 16: En un loop automatico (por ejemplo, 5 iteraciones o indefinido):
	//   - Crear una telemetria.Lectura con SensorID = nombre, Temperatura aleatoria (20-30),
	//     y Timestamp = time.Now().Unix()
	//   - Llamar LlamarRPC(servidorAddr, "Telemetria.RegistrarLectura", lectura, &resp)
	//   - Loguear la respuesta o el error definitivo
	//   - Esperar 3 segundos entre lecturas
	//   - Cada 15 segundos (o cada 5 iteraciones), llamar ObtenerUltimaLectura con
	//     ConsultaUltimaLectura{nombre} para verificar que la lectura quedo registrada
	//
	// IMPORTANTE: usar jsonrpc.Dial en cada llamada o mantener una conexion abierta.
	// Si se mantiene abierta, considerar reconectar ante error.
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	iter := 0

	for {
		iter++
		lectura := telemetria.Lectura{
			SensorID:    nombre,
			Temperatura: 20 + rng.Float64()*10,
			Timestamp:   time.Now().Unix(),
		}

		var resp telemetria.RespuestaLectura
		if err := LlamarRPC(servidorAddr, "Telemetria.RegistrarLectura", lectura, &resp); err != nil {
			fmt.Printf("[CLIENTE %s] Error RPC: %v\n", nombre, err)
		} else {
			fmt.Printf("[CLIENTE %s] Lectura registrada: id=%d msg=%s\n", nombre, resp.ID, resp.Mensaje)
		}

		if iter%5 == 0 {
			consulta := protocolo.ConsultaUltimaLectura{SensorID: nombre}
			var ultima telemetria.Lectura
			if err := LlamarRPC(servidorAddr, "Telemetria.ObtenerUltimaLectura", consulta, &ultima); err != nil {
				fmt.Printf("[CLIENTE %s] Error obteniendo ultima lectura: %v\n", nombre, err)
			} else {
				fmt.Printf("[CLIENTE %s] Ultima lectura: temp=%.2f ts=%d\n", nombre, ultima.Temperatura, ultima.Timestamp)
			}
		}

		time.Sleep(3 * time.Second)
	}
}
