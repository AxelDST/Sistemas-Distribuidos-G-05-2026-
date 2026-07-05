package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"

	"sd-datastore/pkg/gossip"
	"sd-datastore/pkg/replicacion"
)

var (
	idNodo     string
	puertoHTTP string
	puertoRPC  string
	pares      []string
	gossipNodo *gossip.NodoGossip
	servicioQ  *replicacion.ServicioQuorum
	configQ    replicacion.QuorumConfig
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

	// TODO 12: Parsear QUORUM_N, QUORUM_W, QUORUM_R de las variables de entorno.
	// Valores por defecto: N=3, W=2, R=2.
	configQ = replicacion.QuorumConfig{
		N: parsearEntero(os.Getenv("QUORUM_N"), 3),
		W: parsearEntero(os.Getenv("QUORUM_W"), 2),
		R: parsearEntero(os.Getenv("QUORUM_R"), 2),
	}
	fmt.Printf("[NODO %s] QuorumConfig: N=%d W=%d R=%d (válido=%v)\n",
		idNodo, configQ.N, configQ.W, configQ.R, configQ.Validar())

	idNum, _ := strconv.Atoi(idNodo)
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	miDireccionRPC := fmt.Sprintf("%s:%s", hostname, puertoRPC)

	// Inicializar gossip
	gossipNodo = gossip.NuevoNodo(idNum, miDireccionRPC)

	seed := os.Getenv("SEED")
	if seed != "" {
		gossipNodo.Unirse(seed)
	}

	// TODO 13: Inicializar Store, ServicioQuorum y QuorumConfig.
	store := replicacion.NuevoStore()
	servicioQ = &replicacion.ServicioQuorum{
		NodoID: idNodo,
		Store:  store,
		Pares:  pares,
		Config: configQ,
	}

	// Endpoints HTTP
	http.HandleFunc("/estado", manejadorEstado)
	http.HandleFunc("/datos/", manejadorDatos)

	// Servicio RPC
	go iniciarRPC()

	// Loop anti-entropia
	go bucleAntiEntropia()

	addr := ":" + puertoHTTP
	fmt.Printf("[NODO %s] Escuchando HTTP en %s, RPC en %s\n", idNodo, addr, puertoRPC)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// parsearEntero convierte un string a int; devuelve el valor por defecto si falla.
func parsearEntero(s string, defecto int) int {
	if s == "" {
		return defecto
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defecto
	}
	return v
}

// TODO 12b: Implementar parsearPares (usen de PG3).
// Convierte "1=host:port,2=host:port,..." en []string con direcciones RPC.
func parsearPares(peersEnv string) []string {
	if peersEnv == "" {
		return nil
	}
	var resultado []string
	for _, token := range strings.Split(peersEnv, ",") {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		// Formato esperado: "ID=host:port"
		partes := strings.SplitN(token, "=", 2)
		if len(partes) == 2 {
			resultado = append(resultado, partes[1])
		}
	}
	return resultado
}

// TODO 14: Implementar iniciarRPC.
// Crear listener TCP, registrar ServicioGossip y ServicioQuorum, atender conexiones.
func iniciarRPC() {
	srv := rpc.NewServer()

	sgossip := &gossip.ServicioGossip{Nodo: gossipNodo}
	if err := srv.RegisterName("ServicioGossip", sgossip); err != nil {
		log.Fatalf("[NODO %s] Error registrando ServicioGossip: %v", idNodo, err)
	}
	if err := srv.RegisterName("ServicioQuorum", servicioQ); err != nil {
		log.Fatalf("[NODO %s] Error registrando ServicioQuorum: %v", idNodo, err)
	}

	ln, err := net.Listen("tcp", ":"+puertoRPC)
	if err != nil {
		log.Fatalf("[NODO %s] Error abriendo RPC en puerto %s: %v", idNodo, puertoRPC, err)
	}
	fmt.Printf("[NODO %s] RPC escuchando en :%s\n", idNodo, puertoRPC)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[NODO %s] Error aceptando conexión RPC: %v", idNodo, err)
			continue
		}
		go srv.ServeConn(conn)
	}
}

// TODO 15: Implementar bucleAntiEntropia.
// Cada 5 segundos obtener un par con gossipNodo.AntiEntropia(),
// conectarse via RPC, intercambiar miembros y fusionar.
func bucleAntiEntropia() {
	for {
		time.Sleep(5 * time.Second)

		par := gossipNodo.AntiEntropia()
		if par == "" {
			continue
		}

		cliente, err := rpc.Dial("tcp", par)
		if err != nil {
			log.Printf("[NODO %s] AntiEntropia: no pude conectar a %s: %v", idNodo, par, err)
			continue
		}

		req := gossip.Intercambio{
			Remitente: gossipNodo.Direccion,
			Miembros:  gossipNodo.ObtenerMiembros(),
		}
		var resp gossip.Intercambio
		if err := cliente.Call("ServicioGossip.Intercambiar", req, &resp); err != nil {
			log.Printf("[NODO %s] AntiEntropia: error RPC con %s: %v", idNodo, par, err)
			cliente.Close()
			continue
		}
		cliente.Close()

		gossipNodo.FusionarMiembros(resp.Miembros)
		log.Printf("[NODO %s] AntiEntropia: miembros conocidos=%d", idNodo, len(gossipNodo.ObtenerMiembros()))
	}
}

// manejadorEstado responde GET /estado con información del nodo.
func manejadorEstado(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"node_id":  idNodo,
		"miembros": gossipNodo.ObtenerMiembros(),
		"quorum": map[string]int{
			"N": configQ.N,
			"W": configQ.W,
			"R": configQ.R,
		},
		"pares": pares,
	})
}

// manejadorDatos maneja PUT /datos/{clave} y GET /datos/{clave}.
//
//   - PUT: genera timestamp, coordina escritura con quórum W.
//     Responde 200 OK o 503 si no se alcanza quórum.
//   - GET: coordina lectura con quórum R.
//     Responde 200 con el valor, 404 si la clave no existe, o 503 si no hay quórum.
func manejadorDatos(w http.ResponseWriter, r *http.Request) {
	partes := strings.Split(strings.TrimPrefix(r.URL.Path, "/datos/"), "/")
	if len(partes) == 0 || partes[0] == "" {
		http.Error(w, "falta clave", http.StatusBadRequest)
		return
	}
	clave := partes[0]

	switch r.Method {
	case http.MethodPut:
		var body struct {
			Valor string `json:"valor"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ts := time.Now().UnixNano()

		// Escribir localmente primero
		servicioQ.Store.Escribir(clave, body.Valor, ts)

		// Coordinar con los pares para alcanzar quórum W-1 (ya tenemos 1 confirmación local)
		confirmacionesPares := 0
		if len(pares) > 0 {
			if replicacion.CoordinarEscritura(clave, body.Valor, ts, pares, configQ.W-1) {
				confirmacionesPares = configQ.W - 1
			}
		}

		// Verificar quórum total (1 local + confirmaciones de pares)
		totalConfirmaciones := 1 + confirmacionesPares
		if totalConfirmaciones < configQ.W {
			http.Error(w, "quorum de escritura no alcanzado", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"clave":         clave,
			"valor":         body.Valor,
			"timestamp":     ts,
			"confirmaciones": totalConfirmaciones,
		})
		log.Printf("[NODO %s] PUT %s=%s (ts=%d, confirmaciones=%d)", idNodo, clave, body.Valor, ts, totalConfirmaciones)

	case http.MethodGet:
		// Leer localmente
		valorLocal, tsLocal, encontradoLocal := servicioQ.Store.Leer(clave)

		// Coordinar con los pares para alcanzar quórum R
		valorFinal, tsFinal, encontradoFinal := valorLocal, tsLocal, encontradoLocal

		if len(pares) > 0 {
			valorPar, tsPar, okPar := replicacion.CoordinarLectura(clave, pares, configQ.R-1)
			if !okPar && !encontradoLocal {
				// No hay quórum ni localmente ni con pares
				http.Error(w, "quorum de lectura no alcanzado", http.StatusServiceUnavailable)
				return
			}
			// Elegir el valor más reciente entre local y pares
			if okPar && tsPar > tsFinal {
				valorFinal = valorPar
				tsFinal = tsPar
				encontradoFinal = true
			}
		} else if !encontradoLocal {
			http.Error(w, "quorum de lectura no alcanzado", http.StatusServiceUnavailable)
			return
		}

		if !encontradoFinal {
			http.Error(w, "clave no encontrada", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"clave":     clave,
			"valor":     valorFinal,
			"timestamp": tsFinal,
		})
		log.Printf("[NODO %s] GET %s=%s (ts=%d)", idNodo, clave, valorFinal, tsFinal)

	default:
		http.Error(w, "metodo no soportado", http.StatusMethodNotAllowed)
	}
}
