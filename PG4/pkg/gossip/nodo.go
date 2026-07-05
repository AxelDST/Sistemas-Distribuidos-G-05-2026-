package gossip

import (
	"math/rand"
	"sync"
)

// NodoGossip mantiene la membresía del cluster via protocolo gossip.
// Campos:
//   - ID        int    – identificador lógico del nodo
//   - Direccion string – "host:port" RPC con el que este nodo se identifica
//   - Miembros  map[string]bool – conjunto de direcciones conocidas (incluido sí mismo)
//   - mu        sync.RWMutex – protege Miembros
type NodoGossip struct {
	ID        int
	Direccion string
	Miembros  map[string]bool
	mu        sync.RWMutex
}

// NuevoNodo crea un NodoGossip con el ID y dirección dados.
// Inicializa Miembros con solo la dirección propia.
func NuevoNodo(id int, direccion string) *NodoGossip {
	return &NodoGossip{
		ID:        id,
		Direccion: direccion,
		Miembros:  map[string]bool{direccion: true},
	}
}

// Unirse agrega una dirección a la lista de miembros (thread-safe).
func (n *NodoGossip) Unirse(direccion string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Miembros[direccion] = true
}

// AntiEntropia devuelve la dirección de un par aleatorio distinto de sí mismo.
// Retorna "" si no hay otros miembros conocidos.
func (n *NodoGossip) AntiEntropia() string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	otros := make([]string, 0, len(n.Miembros))
	for dir := range n.Miembros {
		if dir != n.Direccion {
			otros = append(otros, dir)
		}
	}
	if len(otros) == 0 {
		return ""
	}
	return otros[rand.Intn(len(otros))]
}

// ObtenerMiembros devuelve una copia del slice de direcciones conocidas (thread-safe).
func (n *NodoGossip) ObtenerMiembros() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	lista := make([]string, 0, len(n.Miembros))
	for dir := range n.Miembros {
		lista = append(lista, dir)
	}
	return lista
}

// FusionarMiembros agrega un slice de direcciones al conjunto de miembros (thread-safe).
func (n *NodoGossip) FusionarMiembros(nuevos []string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, dir := range nuevos {
		if dir != "" {
			n.Miembros[dir] = true
		}
	}
}

// ---------------------------------------------------------------------------
// Tipos y servicio RPC
// ---------------------------------------------------------------------------

// Intercambio es la estructura usada en el RPC de anti-entropía (push-pull).
type Intercambio struct {
	Remitente string
	Miembros  []string
}

// ArgsVacio es un argumento RPC vacío (no usado en PG4, pero se mantiene
// para compatibilidad con el paquete gossip original de PG3).
type ArgsVacio struct{}

// RespIdentificacion devuelve la dirección e ID lógico del nodo consultado.
// No se usa en PG4 (era específico del anillo Chord de PG3), pero se mantiene
// para no romper la API del paquete.
type RespIdentificacion struct {
	Direccion string
	ID        int
}

// ServicioGossip expone los métodos RPC del protocolo Gossip.
type ServicioGossip struct {
	Nodo *NodoGossip
}

// Identificarse devuelve la dirección e ID lógico del nodo (no usado en PG4).
func (s *ServicioGossip) Identificarse(_ ArgsVacio, resp *RespIdentificacion) error {
	resp.Direccion = s.Nodo.Direccion
	resp.ID = s.Nodo.ID
	return nil
}

// Intercambiar implementa la anti-entropía push-pull:
// recibe los miembros del remitente, los fusiona y devuelve los propios.
func (s *ServicioGossip) Intercambiar(req Intercambio, resp *Intercambio) error {
	s.Nodo.FusionarMiembros(req.Miembros)
	if req.Remitente != "" {
		s.Nodo.Unirse(req.Remitente)
	}
	resp.Remitente = s.Nodo.Direccion
	resp.Miembros = s.Nodo.ObtenerMiembros()
	return nil
}
