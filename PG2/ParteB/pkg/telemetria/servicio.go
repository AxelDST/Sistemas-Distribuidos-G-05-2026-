package telemetria

import (
	"fmt"
	"sync"
	"time"

	"sd-comunicacion/pkg/protocolo"
)

// Types compartidos con el protocolo JSON.
type Lectura = protocolo.Lectura
type RespuestaLectura = protocolo.RespuestaLectura
type ConsultaUltimaLectura = protocolo.ConsultaUltimaLectura

// TODO 1: Definir el struct Telemetria que sera el servicio RPC.
// Debe contener un mapa protegido por sync.Mutex para almacenar
// la ultima lectura de cada sensor.
// Sugerencia: usar map[string]Lectura
//
// import "sync" cuando lo necesites

type Telemetria struct {
	mu       sync.Mutex
	lecturas map[string]Lectura
	nextID   int
}

func NuevaTelemetria() *Telemetria {
	return &Telemetria{lecturas: make(map[string]Lectura)}
}

// TODO 2: Implementar el metodo RPC RegistrarLectura.
// Firma requerida por net/rpc:
//   func (t *Telemetria) RegistrarLectura(args Lectura, resp *RespuestaLectura) error
// Debe:
//   - Guardar la lectura en el mapa (protegiendo con mutex)
//   - Asignar un ID incremental a la respuesta
//   - Loguear la lectura recibida (import "fmt" y "time")
//   - Retornar nil en caso de exito

func (t *Telemetria) RegistrarLectura(args Lectura, resp *RespuestaLectura) error {
	if t.lecturas == nil {
		t.lecturas = make(map[string]Lectura)
	}

	t.mu.Lock()
	t.nextID++
	resp.ID = t.nextID
	resp.Mensaje = "ok"
	t.lecturas[args.SensorID] = args
	t.mu.Unlock()

	ts := time.Unix(args.Timestamp, 0).Format(time.RFC3339)
	fmt.Printf("[RPC] Lectura %d sensor=%s temp=%.2f ts=%s\n", resp.ID, args.SensorID, args.Temperatura, ts)

	return nil
}

// TODO 3: Implementar el metodo RPC ObtenerUltimaLectura.
// Firma requerida por net/rpc:
//   func (t *Telemetria) ObtenerUltimaLectura(args ConsultaUltimaLectura, resp *Lectura) error
// Debe:
//   - Buscar en el mapa la ultima lectura del SensorID solicitado
//   - Si no existe, retornar un error con fmt.Errorf
//   - Si existe, copiar el valor a resp y retornar nil

func (t *Telemetria) ObtenerUltimaLectura(args ConsultaUltimaLectura, resp *Lectura) error {
	t.mu.Lock()
	lectura, ok := t.lecturas[args.SensorID]
	t.mu.Unlock()

	if !ok {
		return fmt.Errorf("sin lectura para sensor %s", args.SensorID)
	}

	*resp = lectura
	return nil
}
