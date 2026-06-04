.PHONY: build run-servidor run-cliente test clean docker-up docker-down docker-cliente1 docker-cliente2 docker-logs fmt

# Compilacion local
build:
	go build -o bin/servidor ./cmd/servidor
	go build -o bin/cliente ./cmd/cliente

run-servidor:
	go run ./cmd/servidor

run-cliente:
	NOMBRE=cliente-local SERVIDOR=localhost:1234 go run ./cmd/cliente

# Tests
test:
	go test -v ./...

race:
	go test -race ./...

# Docker: servidor en background, clientes interactivos uno a uno
docker-up:
	docker compose up --build -d servidor

docker-down:
	docker compose down

docker-cliente1:
	docker compose run --rm -it cliente1

docker-cliente2:
	docker compose run --rm -it cliente2

docker-logs:
	docker compose logs -f servidor

# Limpieza
clean:
	rm -rf bin/
	docker compose down -v

# Formato
fmt:
	gofmt -w .
