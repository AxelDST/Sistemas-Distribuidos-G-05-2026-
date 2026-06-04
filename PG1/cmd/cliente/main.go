package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"sd-broadcast/pkg/protocolo"
)

func main() {
	direccionServidor := os.Getenv("SERVIDOR")
	if direccionServidor == "" {
		direccionServidor = "localhost:4000"
	}

	nombre := os.Getenv("NOMBRE")
	if nombre == "" {
		fmt.Print("Ingrese su nombre: ")
		lector := bufio.NewReader(os.Stdin)
		nombreBytes, _, _ := lector.ReadLine()
		nombre = string(nombreBytes)
	}

	conexion, err := net.Dial("tcp", direccionServidor)
	if err != nil {
		log.Fatalf("No se pudo conectar al servidor: %v", err)
	}
	defer conexion.Close()

	err = protocolo.Codificar(conexion, protocolo.NuevoMensaje(nombre, "", "identificacion"))
	if err != nil {
		log.Fatalf("No se pudo enviar identificacion: %v", err)
	}

	go recibirMensajes(conexion)

	lector := bufio.NewReader(os.Stdin)
	for {
		linea, err := lector.ReadString('\n')
		if err != nil {
			log.Printf("Error leyendo stdin: %v", err)
			break
		}

		contenido := strings.TrimSpace(linea)
		if contenido == "" {
			continue
		}

		mensaje := protocolo.NuevoMensaje(nombre, contenido, "broadcast")
		if err := protocolo.Codificar(conexion, mensaje); err != nil {
			log.Printf("Error enviando mensaje: %v", err)
			break
		}
	}

	log.Println("Cliente finalizado")
}

// recibirMensajes lee continuamente desde la conexión e imprime en consola
func recibirMensajes(conexion net.Conn) {
	for {
		mensaje, err := protocolo.Decodificar(conexion)
		if err != nil {
			log.Println("Desconectado del servidor")
			return
		}
		fmt.Printf("[%s] %s: %s\n", mensaje.Timestamp.Format("15:04:05"), mensaje.Emisor, mensaje.Contenido)
	}
}
