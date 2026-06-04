package nodo

import (
	"os"
	"time"
)

// Configuracion representa los parámetros del nodo IoT.
type Configuracion struct {
	ID                string
	Edificio          string
	Aula              string
	BrokerMQTT        string
	IntervaloSegundos time.Duration
}

// CargarConfiguracion lee variables de entorno o usa valores por defecto.
func CargarConfiguracion() Configuracion {
	id := obtenerEnv("NODO_ID", "nodo-01")
	edificio := obtenerEnv("NODO_EDIFICIO", "ingenieria")
	aula := obtenerEnv("NODO_AULA", "lab3")
	broker := obtenerEnv("MQTT_BROKER", "localhost:1883")
	intervalo := obtenerEnv("INTERVALO_SEGUNDOS", "5")

	// TODO: validar que ID, Edificio y Aula no estén vacíos. 
	// Validar que el intervalo sea un número positivo. Si no salir con error.
	// Sugerencia: usar regexp para permitir solo letras, números y guiones.

	duracion, err := time.ParseDuration(intervalo + "s")
	if err != nil {
		duracion = 5 * time.Second
	}

	return Configuracion{
		ID:                id,
		Edificio:          edificio,
		Aula:              aula,
		BrokerMQTT:        broker,
		IntervaloSegundos: duracion,
	}
}

func obtenerEnv(clave, valorPorDefecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return valorPorDefecto
}
