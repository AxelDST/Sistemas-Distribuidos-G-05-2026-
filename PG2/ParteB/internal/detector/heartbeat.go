package detector

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"sd-comunicacion/pkg/protocolo"
)

// TODO 5-8: Implementar envio y recepcion de heartbeats UDP.
// Necesitaras importar:
//   "encoding/json"
//   "fmt"
//   "net"
//   "time"
//   "sd-comunicacion/pkg/protocolo"

// Enviador se encarga de enviar heartbeats UDP periodicamente
type Enviador struct {
	destino   string
	intervalo time.Duration // TODO: usar time.Duration en vez de int64
	nodoID    string
	contador  int
}

// TODO 5: Implementar la funcion NuevaEnviador.
// Debe recibir destino (string), intervalo (time.Duration) y nodoID (string).

func NuevaEnviador(destino string, intervalo time.Duration, nodoID string) *Enviador {
	return &Enviador{
		destino:   destino,
		intervalo: intervalo,
		nodoID:    nodoID,
	}
}

// TODO 6: Implementar el metodo (e *Enviador) Iniciar().
// Debe enviar Heartbeat cada 'intervalo' por UDP al destino configurado.

func (e *Enviador) Iniciar() {
	addr, err := net.ResolveUDPAddr("udp", e.destino)
	if err != nil {
		fmt.Printf("[HEARTBEAT] Error resolviendo destino %s: %v\n", e.destino, err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Printf("[HEARTBEAT] Error abriendo UDP a %s: %v\n", e.destino, err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(e.intervalo)
	defer ticker.Stop()

	for range ticker.C {
		e.contador++
		hb := protocolo.Heartbeat{
			NodoID:    e.nodoID,
			Timestamp: time.Now().Unix(),
			Contador:  e.contador,
		}
		payload, err := json.Marshal(hb)
		if err != nil {
			fmt.Printf("[HEARTBEAT] Error serializando heartbeat: %v\n", err)
			continue
		}
		if _, err := conn.Write(payload); err != nil {
			fmt.Printf("[HEARTBEAT] Error enviando a %s: %v\n", e.destino, err)
		}
	}
}

// Receptor escucha heartbeats y detecta si dejan de llegar.
// Debe manejar estados: alive -> suspect -> dead.
type Receptor struct {
	puerto  string
	timeout time.Duration // TODO: usar time.Duration en vez de int64
	// ultimo debe guardar time.Time o timestamp del ultimo heartbeat recibido
	ultimo time.Time
	// estado puede ser "alive", "suspect" o "dead"
	estado string
	mu     sync.Mutex
}

// TODO 7: Implementar la funcion NuevoReceptor.
// Debe recibir puerto (string) y timeout (time.Duration).

func NuevoReceptor(puerto string, timeout time.Duration) *Receptor {
	return &Receptor{
		puerto:  puerto,
		timeout: timeout,
		ultimo:  time.Now(),
		estado:  "suspect",
	}
}

// TODO 8: Implementar el metodo (r *Receptor) Escuchar().
// Debe:
//   - Escuchar UDP en 'puerto'
//   - Decodificar mensajes JSON tipo protocolo.Heartbeat
//   - Actualizar ultimo timestamp al recibir
//   - En una goroutine separada, revisar periodicamente:
//       si time.Since(ultimo) > timeout: pasar a "suspect"
//       (opcional) si time.Since(ultimo) > 2*timeout: pasar a "dead"
//   - Imprimir cambios de estado por consola

func (r *Receptor) Escuchar() {
	addr, err := net.ResolveUDPAddr("udp", r.puerto)
	if err != nil {
		fmt.Printf("[HEARTBEAT] Error resolviendo puerto %s: %v\n", r.puerto, err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("[HEARTBEAT] Error escuchando UDP en %s: %v\n", r.puerto, err)
		return
	}
	defer conn.Close()

	go r.monitorEstados()

	buf := make([]byte, 2048)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("[HEARTBEAT] Error leyendo UDP: %v\n", err)
			continue
		}

		var hb protocolo.Heartbeat
		if err := json.Unmarshal(buf[:n], &hb); err != nil {
			fmt.Printf("[HEARTBEAT] Error decodificando JSON: %v\n", err)
			continue
		}

		r.mu.Lock()
		r.ultimo = time.Now()
		if r.estado != "alive" {
			r.estado = "alive"
			fmt.Printf("[HEARTBEAT] Estado -> alive (nodo %s)\n", hb.NodoID)
		}
		r.mu.Unlock()
	}
}

func (r *Receptor) monitorEstados() {
	ticker := time.NewTicker(r.timeout / 2)
	defer ticker.Stop()

	for range ticker.C {
		r.mu.Lock()
		sin := time.Since(r.ultimo)
		next := "alive"
		if sin > 2*r.timeout {
			next = "dead"
		} else if sin > r.timeout {
			next = "suspect"
		}
		if next != r.estado {
			r.estado = next
			fmt.Printf("[HEARTBEAT] Estado -> %s (sin heartbeats %.1fs)\n", next, sin.Seconds())
		}
		r.mu.Unlock()
	}
}
