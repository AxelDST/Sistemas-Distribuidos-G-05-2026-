.PHONY: run build test broker-up broker-down docker-nodo1 docker-nodo2 docker-logs clean fmt

run:
	go run ./cmd/nodo

build:
	go build -o bin/nodo ./cmd/nodo

test:
	go test ./...

# Broker
broker-up:
	docker compose up -d nanomq

broker-down:
	docker compose down

# Nodos (uno a uno)
docker-nodo1:
	docker compose run --rm -it nodo1

docker-nodo2:
	docker compose run --rm -it nodo2

# Logs
docker-logs:
	docker compose logs -f

# Limpieza
clean:
	rm -rf bin/
	docker compose down -v

# Formato
fmt:
	gofmt -w .
