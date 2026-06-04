package coap

import (
	"bytes"
	"encoding/json"
	"log"
	"sync"
	"time"

	"sd-iot/pkg/nodo"
	"sd-iot/pkg/sensor"

	"github.com/plgd-dev/go-coap/v3"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
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
//
//	6a. Crear router con mux.NewRouter().
//	6b. Registrar handler GET /temperatura que devuelva JSON con la última lectura.
//	    El JSON debe incluir: nodo_id, temperatura, unidad, timestamp.
//	6c. Registrar handler PUT /config que actualice s.modo y otros parámetros desde el body JSON.
//	6d. Registrar handler GET /config que devuelva la configuración actual en JSON.
//	6e. Llamar coap.ListenAndServe("udp", ":5683", router).
//
// Necesitarás importar:
//
//	"bytes"
//	"encoding/json"
//	"log"
//	"time"
//	"github.com/plgd-dev/go-coap/v3"
//	"github.com/plgd-dev/go-coap/v3/message"
//	"github.com/plgd-dev/go-coap/v3/message/codes"
//	"github.com/plgd-dev/go-coap/v3/mux"
func (s *ServidorCoAP) Iniciar() {
	router := mux.NewRouter()

	router.HandleFunc("/temperatura", func(w mux.ResponseWriter, r *mux.Message) {
		if r.Code() != codes.GET {
			_ = w.SetResponse(codes.MethodNotAllowed, message.TextPlain, bytes.NewReader([]byte("metodo no permitido")))
			return
		}

		lectura := s.sim.ObtenerUltima()
		payload := struct {
			NodoID      string  `json:"nodo_id"`
			Temperatura float64 `json:"temperatura"`
			Unidad      string  `json:"unidad"`
			Timestamp   string  `json:"timestamp"`
		}{
			NodoID:      s.config.ID,
			Temperatura: lectura,
			Unidad:      "C",
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}
		data, err := json.Marshal(payload)
		if err != nil {
			_ = w.SetResponse(codes.InternalServerError, message.TextPlain, bytes.NewReader([]byte("error")))
			return
		}
		_ = w.SetResponse(codes.Content, message.AppJSON, bytes.NewReader(data))
	})

	router.HandleFunc("/config", func(w mux.ResponseWriter, r *mux.Message) {
		switch r.Code() {
		case codes.PUT:
			var req struct {
				Modo             string `json:"modo"`
				IntervaloSeconds int    `json:"intervalo_segundos"`
			}
			body, err := r.ReadBody()
			if err != nil {
				_ = w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte("body invalido")))
				return
			}
			if err := json.Unmarshal(body, &req); err != nil {
				_ = w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte("json invalido")))
				return
			}

			s.mu.Lock()
			if req.Modo != "" {
				s.modo = req.Modo
			}
			if req.IntervaloSeconds > 0 {
				s.config.IntervaloSegundos = time.Duration(req.IntervaloSeconds) * time.Second
			}
			s.mu.Unlock()

			_ = w.SetResponse(codes.Changed, message.TextPlain, bytes.NewReader([]byte("ok")))
			return
		case codes.GET:
			// continue below
		default:
			_ = w.SetResponse(codes.MethodNotAllowed, message.TextPlain, bytes.NewReader([]byte("metodo no permitido")))
			return
		}

		s.mu.RLock()
		payload := struct {
			NodoID           string `json:"nodo_id"`
			Edificio         string `json:"edificio"`
			Aula             string `json:"aula"`
			BrokerMQTT       string `json:"broker_mqtt"`
			IntervaloSeconds int    `json:"intervalo_segundos"`
			Modo             string `json:"modo"`
		}{
			NodoID:           s.config.ID,
			Edificio:         s.config.Edificio,
			Aula:             s.config.Aula,
			BrokerMQTT:       s.config.BrokerMQTT,
			IntervaloSeconds: int(s.config.IntervaloSegundos / time.Second),
			Modo:             s.modo,
		}
		s.mu.RUnlock()

		data, err := json.Marshal(payload)
		if err != nil {
			_ = w.SetResponse(codes.InternalServerError, message.TextPlain, bytes.NewReader([]byte("error")))
			return
		}
		_ = w.SetResponse(codes.Content, message.AppJSON, bytes.NewReader(data))
	})

	log.Printf("Servidor CoAP escuchando en :5683")
	if err := coap.ListenAndServe("udp", ":5683", router); err != nil {
		log.Printf("Error CoAP: %v", err)
	}
}
