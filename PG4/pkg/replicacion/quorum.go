package replicacion

import (
	"fmt"
	"net/rpc"
	"sync"
)

// Estructuras de mensajes RPC.
// Campos necesarios:
//   ArgsEscritura: Clave, Valor, Timestamp
//   RespEscritura: Exito, NodoID
//   ArgsLectura:   Clave
//   RespLectura:   Valor, Timestamp, NodoID

// ---------------------------------------------------------------------------
// Mensajes RPC
// ---------------------------------------------------------------------------

// ArgsEscritura contiene los argumentos para una operación de escritura RPC.
type ArgsEscritura struct {
	Clave     string
	Valor     string
	Timestamp int64
}

// RespEscritura contiene la respuesta de una operación de escritura RPC.
type RespEscritura struct {
	Exito  bool
	NodoID string
}

// ArgsLectura contiene los argumentos para una operación de lectura RPC.
type ArgsLectura struct {
	Clave string
}

// RespLectura contiene la respuesta de una operación de lectura RPC.
type RespLectura struct {
	Valor      string
	Timestamp  int64
	NodoID     string
	Encontrado bool
}

// ---------------------------------------------------------------------------
// TODO 1: Definir QuorumConfig con N, W, R.
// Agregar metodo Validar() bool que retorne W+R > N.

// QuorumConfig define los parámetros N, W, R del quórum.
//   - N: número total de réplicas
//   - W: quórum de escritura (mínimo de confirmaciones necesarias)
//   - R: quórum de lectura (mínimo de respuestas necesarias)
type QuorumConfig struct {
	N int
	W int
	R int
}

// Validar retorna true si la configuración de quórum es válida (W+R > N).
// Esta condición garantiza que toda lectura solapa con toda escritura.
func (c QuorumConfig) Validar() bool {
	return c.W+c.R > c.N
}

// ---------------------------------------------------------------------------
// Todo 2: Store es el almacenamiento local con timestamps.

// entrada representa un valor almacenado junto a su timestamp.
type entrada struct {
	Valor     string
	Timestamp int64
}

// Store es el almacenamiento local key-value con control de versiones por timestamp.
type Store struct {
	mu   sync.RWMutex
	data map[string]entrada
}

// TODO 3: Implementar NuevoStore.
func NuevoStore() *Store {
	return &Store{
		data: make(map[string]entrada),
	}
}

// TODO 4: Implementar Escribir.
// Si el timestamp recibido es mayor o igual al almacenado, actualizar.
// Retornar true si se actualizo, false si se ignoro.
func (s *Store) Escribir(clave, valor string, timestamp int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	actual, existe := s.data[clave]
	if !existe || timestamp >= actual.Timestamp {
		s.data[clave] = entrada{Valor: valor, Timestamp: timestamp}
		return true
	}
	return false
}

// TODO 5: Implementar Leer.
// Retornar valor, timestamp y un bool indicando si la clave existe.
func (s *Store) Leer(clave string) (string, int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, existe := s.data[clave]
	if !existe {
		return "", 0, false
	}
	return e.Valor, e.Timestamp, true
}

// TODO 6: Implementar Sincronizar.
// Misma lógica que Escribir (es idempotente).
// Se usa para read-repair.
func (s *Store) Sincronizar(clave, valor string, timestamp int64) bool {
	return s.Escribir(clave, valor, timestamp)
}

// ---------------------------------------------------------------------------
// ServicioQuorum – métodos RPC del nodo réplica
// ---------------------------------------------------------------------------

// ServicioQuorum expone métodos RPC para lecturas y escrituras con quorum.
type ServicioQuorum struct {
	NodoID string
	Store  *Store
	Pares  []string
	Config QuorumConfig
}

// TODO 7: Implementar Escribir (RPC).
// Recibe ArgsEscritura, delega en Store.Escribir, responde con Exito=true y NodoID.
func (s *ServicioQuorum) Escribir(args ArgsEscritura, resp *RespEscritura) error {
	s.Store.Escribir(args.Clave, args.Valor, args.Timestamp)
	resp.Exito = true
	resp.NodoID = s.NodoID
	return nil
}

// TODO 8: Implementar Leer (RPC).
// Recibe ArgsLectura, delega en Store.Leer, responde con valor, timestamp y NodoID.
func (s *ServicioQuorum) Leer(args ArgsLectura, resp *RespLectura) error {
	valor, ts, encontrado := s.Store.Leer(args.Clave)
	resp.Valor = valor
	resp.Timestamp = ts
	resp.NodoID = s.NodoID
	resp.Encontrado = encontrado
	return nil
}

// TODO 9: Implementar Sincronizar (RPC).
// Recibe ArgsEscritura, delega en Store.Sincronizar para read-repair.
func (s *ServicioQuorum) Sincronizar(args ArgsEscritura, resp *RespEscritura) error {
	actualizado := s.Store.Sincronizar(args.Clave, args.Valor, args.Timestamp)
	resp.Exito = actualizado
	resp.NodoID = s.NodoID
	return nil
}

// CoordinarEscritura es la funcion cliente que coordina el quorum de escritura.
// Conecta RPC a cada par, invoca Escribir, y retorna true si W o mas confirmaron.
// TODO 10: Implementar CoordinarEscritura.
func CoordinarEscritura(clave, valor string, timestamp int64, pares []string, w int) bool {
	type resultado struct {
		exito bool
	}

	ch := make(chan resultado, len(pares))

	for _, par := range pares {
		go func(addr string) {
			cliente, err := rpc.Dial("tcp", addr)
			if err != nil {
				ch <- resultado{false}
				return
			}
			defer cliente.Close()

			args := ArgsEscritura{Clave: clave, Valor: valor, Timestamp: timestamp}
			var resp RespEscritura
			err = cliente.Call("ServicioQuorum.Escribir", args, &resp)
			if err != nil || !resp.Exito {
				ch <- resultado{false}
				return
			}
			ch <- resultado{true}
		}(par)
	}

	confirmaciones := 0
	for range pares {
		r := <-ch
		if r.exito {
			confirmaciones++
		}
	}

	return confirmaciones >= w
}

// CoordinarLectura es la funcion cliente que coordina el quorum de lectura.
// Conecta RPC a cada par, invoca Leer, y retorna el valor con el timestamp mas grande.
// Retorna true si obtuvo al menos R respuestas.
// TODO 11: Implementar CoordinarLectura.
func CoordinarLectura(clave string, pares []string, r int) (string, int64, bool) {
	type resultado struct {
		resp RespLectura
		err  error
	}

	ch := make(chan resultado, len(pares))

	for _, par := range pares {
		go func(addr string) {
			cliente, err := rpc.Dial("tcp", addr)
			if err != nil {
				ch <- resultado{err: err}
				return
			}
			defer cliente.Close()

			args := ArgsLectura{Clave: clave}
			var resp RespLectura
			err = cliente.Call("ServicioQuorum.Leer", args, &resp)
			ch <- resultado{resp: resp, err: err}
		}(par)
	}

	// Recolectar respuestas
	respuestas := make([]resultado, 0, len(pares))
	for range pares {
		res := <-ch
		if res.err == nil {
			respuestas = append(respuestas, res)
		}
	}

	if len(respuestas) < r {
		return "", 0, false
	}

	// Encontrar el valor con timestamp más alto
	var mejor RespLectura
	for _, res := range respuestas {
		if res.resp.Encontrado && res.resp.Timestamp > mejor.Timestamp {
			mejor = res.resp
		}
	}

	// Read-repair en background: actualizar réplicas con timestamp menor
	if mejor.Encontrado {
		for i, par := range pares {
			// Determinar si este par necesita reparación
			var tsReplica int64
			if i < len(respuestas) {
				tsReplica = respuestas[i].resp.Timestamp
			}
			if tsReplica < mejor.Timestamp {
				go func(addr string, val RespLectura) {
					cliente, err := rpc.Dial("tcp", addr)
					if err != nil {
						return
					}
					defer cliente.Close()
					args := ArgsEscritura{
						Clave:     clave,
						Valor:     val.Valor,
						Timestamp: val.Timestamp,
					}
					var resp RespEscritura
					_ = cliente.Call("ServicioQuorum.Sincronizar", args, &resp)
					fmt.Printf("[READ-REPAIR] %s ← %s=%s (ts=%d)\n", addr, clave, val.Valor, val.Timestamp)
				}(par, mejor)
			}
		}
	}

	if !mejor.Encontrado {
		return "", 0, false
	}
	return mejor.Valor, mejor.Timestamp, true
}
