package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"sd-descubrimiento/pkg/dht"
	"sd-descubrimiento/pkg/gossip"
)

var (
	idNodo      string
	idNum       int
	puertoHTTP  string
	puertoRPC   string
	pares       map[int]string
	gossipNodo  *gossip.NodoGossip
	chordNodo   *dht.NodoChord
	miDireccion string
)

func main() {
	idNodo = os.Getenv("NODO_ID")
	if idNodo == "" {
		idNodo = "1"
	}
	puertoHTTP = os.Getenv("HTTP_PORT")
	if puertoHTTP == "" {
		puertoHTTP = "8080"
	}
	puertoRPC = os.Getenv("RPC_PORT")
	if puertoRPC == "" {
		puertoRPC = "5000"
	}

	pares = parsearPares(os.Getenv("PEERS"))
	idNum, _ = strconv.Atoi(idNodo)
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	miDireccion = fmt.Sprintf("%s:%s", hostname, puertoRPC)

	// Inicializar nodos
	gossipNodo = gossip.NuevoNodo(idNum, miDireccion)

	seed := os.Getenv("SEED")
	if seed != "" {
		// Resolver la Direccion canónica del seed via Identificarse para
		// evitar duplicados en la membresía. Si falla, usar el seed crudo.
		c, err := rpc.Dial("tcp", seed)
		if err == nil {
			var ri gossip.RespIdentificacion
			if c.Call("ServicioGossip.Identificarse", gossip.ArgsVacio{}, &ri) == nil && ri.Direccion != "" {
				gossipNodo.Unirse(ri.Direccion)
			} else {
				gossipNodo.Unirse(seed)
			}
			c.Close()
		} else {
			gossipNodo.Unirse(seed)
		}
	}

	// TODO 15: inicializar chordNodo con sucesor y predecesor calculados
	// a partir de NODO_ID y PEERS (IDs ordenados en el anillo 0-255).
	chordNodo = inicializarChord()

	// Endpoints HTTP
	http.HandleFunc("/estado", manejadorEstado)
	http.HandleFunc("/almacenar", manejadorAlmacenar)
	http.HandleFunc("/buscar", manejadorBuscar)

	// Servicio RPC
	go iniciarRPC()

	// Loop anti-entropia
	go bucleAntiEntropia()

	// Loop estabilización del anillo Chord (cada 10 segundos)
	go bucleEstabilizacionChord()

	addr := ":" + puertoHTTP
	fmt.Printf("[NODO %s] Escuchando HTTP en %s, RPC en %s\n", idNodo, addr, puertoRPC)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// inicializarChord calcula sucesor y predecesor del nodo actual usando
// NODO_ID y PEERS, ordena los IDs en el anillo 0-255 y construye NodoChord.
func inicializarChord() *dht.NodoChord {
	// Construir mapa completo: propio ID + todos los PEERS
	todos := make(map[int]string)
	todos[idNum] = miDireccion
	for id, dir := range pares {
		todos[id] = dir
		// También añadir los PEERS al Gossip para que la membresía inicial esté completa
		gossipNodo.Unirse(dir)
	}

	// Ordenar IDs en el anillo
	ids := make([]int, 0, len(todos))
	for id := range todos {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	n := len(ids)
	miIdx := 0
	for i, v := range ids {
		if v == idNum {
			miIdx = i
			break
		}
	}

	var sucID, predID int
	var sucDir, predDir string
	if n == 1 {
		// Único nodo: apunta a sí mismo
		sucID = idNum
		sucDir = miDireccion
		predID = idNum
		predDir = miDireccion
	} else {
		sucID = ids[(miIdx+1)%n]
		sucDir = todos[sucID]
		predID = ids[(miIdx-1+n)%n]
		predDir = todos[predID]
	}

	fmt.Printf("[NODO %s] Chord init: anillo=%v pred=%d suc=%d\n", idNodo, ids, predID, sucID)
	return dht.NuevoNodo(idNum, miDireccion, sucDir, sucID, predDir, predID)
}

// parsearPares parsea los valores pasados como argumentos desde el shell
func parsearPares(peersEnv string) map[int]string {
	resultado := make(map[int]string)
	if peersEnv == "" {
		return resultado
	}
	partes := strings.Split(peersEnv, ",")
	for _, p := range partes {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(kv) != 2 {
			continue
		}
		id, err := strconv.Atoi(kv[0])
		if err != nil {
			continue
		}
		resultado[id] = kv[1]
	}
	return resultado
}

// TODO 16: Implementar iniciarRPC.
// Debe crear un listener TCP en puertoRPC, registrar un servicio RPC
// que maneje intercambios de Gossip y lookups de DHT, y atender conexiones.
func iniciarRPC() {
	srv := rpc.NewServer()
	if err := srv.RegisterName("ServicioGossip", &gossip.ServicioGossip{Nodo: gossipNodo}); err != nil {
		log.Fatalf("[NODO %s] Error registrando ServicioGossip: %v", idNodo, err)
	}
	if err := srv.RegisterName("ServicioChord", &dht.ServicioChord{Nodo: chordNodo}); err != nil {
		log.Fatalf("[NODO %s] Error registrando ServicioChord: %v", idNodo, err)
	}

	ln, err := net.Listen("tcp", ":"+puertoRPC)
	if err != nil {
		log.Fatalf("[NODO %s] Error escuchando RPC en puerto %s: %v", idNodo, puertoRPC, err)
	}
	fmt.Printf("[NODO %s] RPC escuchando en :%s\n", idNodo, puertoRPC)
	srv.Accept(ln)
}

// TODO 17: Implementar bucleAntiEntropia.
// Cada 5 segundos, obtener un par con gossipNodo.AntiEntropia().
// Si hay par, conectarse via RPC, intercambiar miembros y fusionar.
func bucleAntiEntropia() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		par := gossipNodo.AntiEntropia()
		if par == "" {
			continue
		}
		c, err := rpc.Dial("tcp", par)
		if err != nil {
			continue
		}
		req := gossip.Intercambio{
			Remitente: gossipNodo.Direccion,
			Miembros:  gossipNodo.ObtenerMiembros(),
		}
		var resp gossip.Intercambio
		if err := c.Call("ServicioGossip.Intercambiar", req, &resp); err == nil {
			gossipNodo.FusionarMiembros(resp.Miembros)
			if resp.Remitente != "" {
				gossipNodo.Unirse(resp.Remitente)
			}
			fmt.Printf("[NODO %s] Gossip anti-entropía con %s → miembros conocidos: %v\n",
				idNodo, par, gossipNodo.ObtenerMiembros())
		}
		c.Close()
	}
}

// TODO 18: Implementar manejadorEstado.
// GET /estado devuelve JSON con node_id, miembros y finger_table.
func manejadorEstado(w http.ResponseWriter, r *http.Request) {
	fingerTable := []map[string]interface{}{}
	if chordNodo != nil {
		for i := 0; i < 3; i++ {
			fingerTable = append(fingerTable, map[string]interface{}{
				"indice":    i,
				"direccion": chordNodo.FingerTable[i],
				"id":        chordNodo.FingerTableIDs[i],
			})
		}
	}

	estado := map[string]interface{}{
		"node_id":      idNodo,
		"direccion":    miDireccion,
		"miembros":     gossipNodo.ObtenerMiembros(),
		"finger_table": fingerTable,
	}
	if chordNodo != nil {
		estado["sucesor"] = map[string]interface{}{
			"direccion": chordNodo.Sucesor,
			"id":        chordNodo.SucesorID,
		}
		estado["predecesor"] = map[string]interface{}{
			"direccion": chordNodo.Predecesor,
			"id":        chordNodo.PredecesorID,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(estado)
}

// TODO 19: Implementar manejadorAlmacenar.
// POST /almacenar recibe JSON {"clave": int, "valor": string}.
// Si el nodo es responsable (chordNodo.EsResponsable), almacenar localmente.
// Si no, reenviar al sucesor via RPC.
func manejadorAlmacenar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if chordNodo == nil {
		http.Error(w, "Chord no inicializado", http.StatusServiceUnavailable)
		return
	}

	var body struct {
		Clave int    `json:"clave"`
		Valor string `json:"valor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	args := dht.ArgsStore{Clave: body.Clave, Valor: body.Valor}
	var resp dht.RespStore

	// Usar el servicio Chord local que hace forwarding automático
	svc := &dht.ServicioChord{Nodo: chordNodo}
	if err := svc.Almacenar(args, &resp); err != nil {
		http.Error(w, fmt.Sprintf("Error almacenando: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":               true,
		"clave":            body.Clave,
		"nodo_responsable": resp.NodoResponsable,
		"nodo_id":          resp.NodoID,
	})
}

// TODO 20: Implementar manejadorBuscar.
// GET /buscar?clave=X
// Si responsable, devolver valor. Si no, reenviar o redirigir.
func manejadorBuscar(w http.ResponseWriter, r *http.Request) {
	if chordNodo == nil {
		http.Error(w, "Chord no inicializado", http.StatusServiceUnavailable)
		return
	}

	claveStr := r.URL.Query().Get("clave")
	if claveStr == "" {
		http.Error(w, "Parámetro 'clave' requerido", http.StatusBadRequest)
		return
	}
	clave, err := strconv.Atoi(claveStr)
	if err != nil {
		http.Error(w, "Clave debe ser un entero", http.StatusBadRequest)
		return
	}

	args := dht.ArgsLookup{Clave: clave}
	var resp dht.RespLookup

	// Usar el servicio Chord local que hace forwarding automático
	svc := &dht.ServicioChord{Nodo: chordNodo}
	if err := svc.Obtener(args, &resp); err != nil {
		http.Error(w, fmt.Sprintf("Error buscando: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clave":            clave,
		"valor":            resp.Valor,
		"encontrado":       resp.Encontrado,
		"nodo_responsable": resp.NodoResponsable,
		"nodo_id":          resp.NodoID,
	})
}

// bucleEstabilizacionChord recalcula el anillo Chord cada 10 segundos.
// Consulta a todos los miembros descubiertos por Gossip via RPC
// (ServicioGossip.Identificarse) para obtener su ID lógico y Direccion
// canónica, luego reordena los IDs y recalcula sucesor/predecesor.
// Si un miembro deja de responder (crash), se excluye del recalculo y
// el anillo se auto-repara en ~10s. Si Gossip descubre un nuevo nodo,
// entra al anillo en la proxima estabilización.
func bucleEstabilizacionChord() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if chordNodo == nil {
			continue
		}
		miembros := gossipNodo.ObtenerMiembros()
		conocidos := map[int]string{idNum: miDireccion}
		for _, dir := range miembros {
			if dir == miDireccion {
				continue
			}
			c, err := rpc.Dial("tcp", dir)
			if err != nil {
				continue
			}
			var ri gossip.RespIdentificacion
			if c.Call("ServicioGossip.Identificarse", gossip.ArgsVacio{}, &ri) == nil {
				conocidos[ri.ID] = ri.Direccion
			}
			c.Close()
		}
		ids := []int{}
		for id := range conocidos {
			ids = append(ids, id)
		}
		sort.Ints(ids)
		n := len(ids)
		miIdx := 0
		for i, v := range ids {
			if v == idNum {
				miIdx = i
				break
			}
		}
		predID := ids[(miIdx-1+n)%n]
		sucID := ids[(miIdx+1)%n]
		if n == 1 || predID == idNum {
			chordNodo.ActualizarAnillo(miDireccion, idNum, miDireccion, idNum)
		} else {
			chordNodo.ActualizarAnillo(conocidos[sucID], sucID, conocidos[predID], predID)
		}
		fmt.Printf("[NODO %s] Estabilización Chord: anillo=%v pred=%d suc=%d\n",
			idNodo, ids, predID, sucID)
	}
}
