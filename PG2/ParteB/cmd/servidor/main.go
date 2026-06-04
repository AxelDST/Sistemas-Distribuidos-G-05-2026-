package main

import (
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strings"
	"time"

	// TODO: importar "net" para Listen
	// TODO: importar "net/rpc/jsonrpc" para servir con JSON
	// TODO: importar "sd-comunicacion/pkg/telemetria"
	// TODO: importar "sd-comunicacion/internal/detector"
	// TODO: importar "time"

	"sd-comunicacion/internal/detector"
	"sd-comunicacion/pkg/telemetria"
)

func main() {
	// Configuracion por variables de entorno (para Docker)
	rpcPuerto := os.Getenv("RPC_PUERTO")
	if rpcPuerto == "" {
		rpcPuerto = ":1234"
	}
	heartbeatPuerto := os.Getenv("HEARTBEAT_PUERTO")
	if heartbeatPuerto == "" {
		heartbeatPuerto = ":4001"
	}
	nodoID := os.Getenv("NODO_ID")
	if nodoID == "" {
		nodoID = "servidor-1"
	}

	fmt.Printf("[SERVIDOR] Iniciando nodo %s\n", nodoID)
	fmt.Printf("[SERVIDOR] RPC en %s | Heartbeat en %s\n", rpcPuerto, heartbeatPuerto)

	// TODO 10: Crear una instancia de telemetria.Telemetria e inicializar su mapa.
	// Sugerencia: usar make(map[string]telemetria.Lectura)
	// Necesitaras: import "sd-comunicacion/pkg/telemetria"
	servicio := telemetria.NuevaTelemetria()

	// TODO 11: Registrar el servicio con rpc.Register(servicio).
	// Si hay error, loguear y salir.
	// Necesitaras: import "net/rpc"
	if err := rpc.Register(servicio); err != nil {
		fmt.Printf("[SERVIDOR] Error registrando RPC: %v\n", err)
		return
	}

	// TODO 12: Iniciar el enviador de heartbeats.
	// Crear un detector.Enviador que envie a los peers conocidos.
	// Necesitaras: import "sd-comunicacion/internal/detector" y "time"
	heartbeatDestino := os.Getenv("HEARTBEAT_DESTINO")
	if heartbeatDestino != "" {
		fmt.Printf("[SERVIDOR] Enviando heartbeats a %s\n", heartbeatDestino)
		destinos := strings.Split(heartbeatDestino, ",")
		for _, destino := range destinos {
			destino = strings.TrimSpace(destino)
			if destino == "" {
				continue
			}
			enviador := detector.NuevaEnviador(destino, 2*time.Second, nodoID)
			go enviador.Iniciar()
		}
	} else {
		fmt.Println("[SERVIDOR] Sin destino de heartbeat configurado (modo standalone)")
	}

	// TODO 13: Escuchar conexiones TCP en rpcPuerto.
	// Usar net.Listen("tcp", rpcPuerto).
	// En un loop infinito:
	//   conn, err := escuchador.Accept()
	//   if err != nil { continue }
	//   go jsonrpc.ServeConn(conn)  // usar jsonrpc, no rpc.ServeConn
	//
	// IMPORTANTE: jsonrpc.ServeConn atiende UNA conexion con protocolo JSON.
	// Para multiples clientes concurrentes, una goroutine por conn.
	listener, err := net.Listen("tcp", rpcPuerto)
	if err != nil {
		fmt.Printf("[SERVIDOR] Error escuchando en %s: %v\n", rpcPuerto, err)
		return
	}
	defer listener.Close()

	fmt.Println("[SERVIDOR] Listo. Esperando conexiones...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go jsonrpc.ServeConn(conn)
	}
}
