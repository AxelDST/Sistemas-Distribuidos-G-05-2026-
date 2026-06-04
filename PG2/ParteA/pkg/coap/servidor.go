package coap

import (
	"sd-iot/pkg/nodo"
	"sd-iot/pkg/sensor"
	"sync"
)

// ServidorCoAP expone recursos REST sobre UDP.
type ServidorCoAP struct {
	sim    *sensor.Simulador
	config nodo.Configuracion
	mu     sync.RWMutex
	modo   string
}

// NuevoServidor crea la instancia del servidor CoAP.
func NuevoServidor(sim *sensor.Simulador, config nodo.Configuracion) *ServidorCoAP {
	return &ServidorCoAP{
		sim:    sim,
		config: config,
		modo:   "automatico",
	}
}

// TODO 6: Iniciar arranca el servidor UDP en el puerto 5683.
// Debe:
//   6a. Crear router con mux.NewRouter().
//   6b. Registrar handler GET /temperatura que devuelva JSON con la última lectura.
//       El JSON debe incluir: nodo_id, temperatura, unidad, timestamp.
//   6c. Registrar handler PUT /config que actualice s.modo y otros parámetros desde el body JSON.
//   6d. Registrar handler GET /config que devuelva la configuración actual en JSON.
//   6e. Llamar coap.ListenAndServe("udp", ":5683", router).
//
// Necesitarás importar:
//   "bytes"
//   "encoding/json"
//   "log"
//   "time"
//   "github.com/plgd-dev/go-coap/v3"
//   "github.com/plgd-dev/go-coap/v3/message"
//   "github.com/plgd-dev/go-coap/v3/message/codes"
//   "github.com/plgd-dev/go-coap/v3/mux"
func (s *ServidorCoAP) Iniciar() {
	// COMPLETAR
}
