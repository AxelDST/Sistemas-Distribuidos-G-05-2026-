package protocolo

// Heartbeat es el mensaje enviado periodicamente por UDP
type Heartbeat struct {
	NodoID    string `json:"nodo_id"`
	Timestamp int64  `json:"timestamp"`
	Contador  int    `json:"contador"`
}

// TODO: Definir struct Lectura para RPC RegistrarLectura.
// Debe tener al menos los campos:
//   - SensorID    string `json:"sensor_id"`
//   - Temperatura float64 `json:"temperatura"`
//   - Timestamp   int64   `json:"timestamp"`
type Lectura struct {
	SensorID    string  `json:"sensor_id"`
	Temperatura float64 `json:"temperatura"`
	Timestamp   int64   `json:"timestamp"`
}

// TODO: Definir struct RespuestaLectura para respuesta de RegistrarLectura.
// Debe tener al menos los campos:
//   - ID      int    `json:"id"`
//   - Mensaje string `json:"mensaje"`
type RespuestaLectura struct {
	ID      int    `json:"id"`
	Mensaje string `json:"mensaje"`
}

// TODO: Definir struct ConsultaUltimaLectura para RPC ObtenerUltimaLectura.
// Debe tener al menos el campo:
//   - SensorID string `json:"sensor_id"`
type ConsultaUltimaLectura struct {
	SensorID string `json:"sensor_id"`
}
