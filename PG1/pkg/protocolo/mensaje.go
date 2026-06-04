package protocolo

import (
	"encoding/json"
	"io"
	"time"
)

// Mensaje representa un mensaje en el protocolo de comunicación
type Mensaje struct {
	Emisor    string    `json:"emisor"`
	Contenido string    `json:"contenido"`
	Tipo      string    `json:"tipo"` // "broadcast", "privado", "sistema", "identificacion"
	Timestamp time.Time `json:"timestamp"`
}

// NuevoMensaje crea un mensaje con el timestamp actual
func NuevoMensaje(emisor, contenido, tipo string) Mensaje {
	return Mensaje{
		Emisor:    emisor,
		Contenido: contenido,
		Tipo:      tipo,
		Timestamp: time.Now(),
	}
}

// Codificar escribe un mensaje como JSON en el escritor
func Codificar(escritor io.Writer, mensaje Mensaje) error {
	codificador := json.NewEncoder(escritor)
	return codificador.Encode(mensaje)
}

// Decodificar lee un mensaje JSON desde el lector
func Decodificar(lector io.Reader) (Mensaje, error) {
	var mensaje Mensaje
	decodificador := json.NewDecoder(lector)
	err := decodificador.Decode(&mensaje)
	return mensaje, err
}
