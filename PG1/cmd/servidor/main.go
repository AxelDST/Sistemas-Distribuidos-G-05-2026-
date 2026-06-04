package main

import (
	"log"
	"net"
	"os"

	"sd-broadcast/internal/registro"
	"sd-broadcast/pkg/protocolo"
)

const puertoPorDefecto = "4000"

func main() {
	puerto := os.Getenv("PUERTO")
	if puerto == "" {
		puerto = puertoPorDefecto
	}

	escuchador, err := net.Listen("tcp", ":"+puerto)
	if err != nil {
		log.Fatalf("No se pudo iniciar el escuchador: %v", err)
	}
	defer escuchador.Close()

	log.Printf("Servidor de broadcast escuchando en :%s", puerto)

	registroClientes := registro.NuevoRegistro()

	// go iniciarDescubrimientoUDP(puerto)

	for {
		conexion, err := escuchador.Accept()
		if err != nil {
			log.Printf("Error al aceptar conexión: %v", err)
			continue
		}

		go manejarCliente(conexion, registroClientes)
	}
}

func manejarCliente(conexion net.Conn, registroClientes *registro.RegistroClientes) {
	defer conexion.Close()

	mensaje, err := protocolo.Decodificar(conexion)
	if err != nil {
		log.Printf("Error al recibir identificación: %v", err)
		return
	}

	nombreCliente := mensaje.Emisor

	log.Printf("Cliente conectado: %s desde %s", nombreCliente, conexion.RemoteAddr())

	registroClientes.Agregar(nombreCliente, conexion)

	difundirMensaje(registroClientes, protocolo.NuevoMensaje("Sistema", nombreCliente+" se unió", "sistema"), nombreCliente)

	defer registroClientes.Eliminar(nombreCliente)
	defer difundirMensaje(registroClientes, protocolo.NuevoMensaje("Sistema", nombreCliente+" se desconectó", "sistema"), nombreCliente)

	for {
		mensaje, err := protocolo.Decodificar(conexion)
		if err != nil {
			break
		}

		if mensaje.Tipo == "broadcast" {
			difundirMensaje(registroClientes, mensaje, nombreCliente)
		}
	}

	log.Printf("Cliente desconectado: %s", nombreCliente)
}

// difundirMensaje envía un mensaje a todos los clientes excepto al emisor indicado
func difundirMensaje(registroClientes *registro.RegistroClientes, mensaje protocolo.Mensaje, exceptoEmisor string) {
	nombres := registroClientes.Nombres()
	conexiones := registroClientes.ObtenerConexiones()
	for i, conn := range conexiones {
		if i < len(nombres) && nombres[i] != exceptoEmisor {
			err := protocolo.Codificar(conn, mensaje)
			if err != nil {
				log.Printf("Error al enviar mensaje a %s: %v", nombres[i], err)
			}
		}
	}
}
