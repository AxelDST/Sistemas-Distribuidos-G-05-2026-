package nodo

import (
	"log"
	"os"
	"regexp"
	"strconv"
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

	validarCampo("NODO_ID", id)
	validarCampo("NODO_EDIFICIO", edificio)
	validarCampo("NODO_AULA", aula)
	segundos, err := strconv.Atoi(intervalo)
	if err != nil || segundos <= 0 {
		log.Fatalf("INTERVALO_SEGUNDOS invalido: %q", intervalo)
	}
	duracion := time.Duration(segundos) * time.Second

	return Configuracion{
		ID:                id,
		Edificio:          edificio,
		Aula:              aula,
		BrokerMQTT:        broker,
		IntervaloSegundos: duracion,
	}
}

func validarCampo(nombre, valor string) {
	if valor == "" {
		log.Fatalf("%s no puede estar vacio", nombre)
	}
	permitido := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !permitido.MatchString(valor) {
		log.Fatalf("%s invalido: %q", nombre, valor)
	}
}

func obtenerEnv(clave, valorPorDefecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return valorPorDefecto
}
