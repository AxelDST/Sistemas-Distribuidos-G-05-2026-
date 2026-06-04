package registro

import (
	"net"
	"sync"
)

// RegistroClientes mantiene el listado de conexiones activas de forma segura
type RegistroClientes struct {
	mu       sync.RWMutex
	clientes map[string]net.Conn
}

// NuevoRegistro crea un registro vacío
func NuevoRegistro() *RegistroClientes {
	return &RegistroClientes{
		clientes: make(map[string]net.Conn),
	}
}

// Agregar añade un cliente al registro
func (r *RegistroClientes) Agregar(nombre string, conexion net.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clientes[nombre] = conexion
}

// Eliminar remueve un cliente del registro
func (r *RegistroClientes) Eliminar(nombre string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clientes, nombre)
}

// ObtenerConexiones devuelve una copia de todas las conexiones activas
func (r *RegistroClientes) ObtenerConexiones() []net.Conn {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conns := make([]net.Conn, 0, len(r.clientes))
	for _, c := range r.clientes {
		conns = append(conns, c)
	}
	return conns
}

// Cantidad devuelve el número de clientes conectados
func (r *RegistroClientes) Cantidad() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clientes)
}

// Nombres devuelve un slice con los nombres de los clientes
func (r *RegistroClientes) Nombres() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n := make([]string, 0, len(r.clientes))
	for name := range r.clientes {
		n = append(n, name)
	}
	return n
}
