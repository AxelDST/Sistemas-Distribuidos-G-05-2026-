# Nodo IoT Smart Campus

Proyecto base para la parte A de la Practica Guiada 2: comunicacion MQTT + CoAP.

## Integrantes

- Apellidos y Nombres 1
- Apellidos y Nombres 2
- Apellidos y Nombres 3

## Ejecucion

### Local

Requisito: broker MQTT corriendo en `localhost:1883` (puede ser NanoMQ local).

```bash
# Terminal 1: Broker (si no tienes uno local)
make broker-up

# Terminal 2: Nodo
make run
```

### Docker Compose (interactivo)

**1. Levantar solo el broker** (en background):
```bash
make broker-up
```

**2. Lanzar nodos** (en terminales separadas):
```bash
# Terminal 2: Nodo 1
make docker-nodo1

# Terminal 3: Nodo 2
make docker-nodo2
```

**3. Ver logs del broker o nodos**:
```bash
make docker-logs
```

**4. Detener todo**:
```bash
make broker-down
```

## Requisitos completados

- [ ] Cliente MQTT con testamento (LWT): `nodo/{id}/estado` -> `{"estado":"offline"}`
- [ ] Publicar estado `{"estado":"online"}` retenido tras conectar
- [ ] Loop de lecturas simuladas cada 5 s en `campus/{edificio}/{aula}/sensor/temperatura` con QoS 1
- [ ] Suscripcion a comandos en `campus/{edificio}/{aula}/actuador/cmd` con accion impresa
- [ ] Servidor CoAP con recursos:
  - [ ] `GET /temperatura` -> ultima lectura en JSON
  - [ ] `PUT /config` -> actualizar configuracion local
  - [ ] `GET /config` -> configuracion actual en JSON
- [ ] Docker Compose con al menos 1 nodo + NanoMQ broker

## Captura de ejecucion

_(Adjuntar log o captura de pantalla con multiples nodos publicando y respondiendo CoAP)_
