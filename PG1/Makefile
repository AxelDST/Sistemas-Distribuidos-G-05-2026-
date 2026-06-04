.PHONY: build run test clean

build:
	go build -o bin/servidor ./cmd/servidor
	go build -o bin/cliente ./cmd/cliente

run-servidor:
	go run ./cmd/servidor

run-cliente:
	go run ./cmd/cliente

test:
	go test -v ./...

test-race:
	go test -race -v ./...

clean:
	rm -rf bin/

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down
