# Nodo Smart Campus
FROM golang:1.22-alpine AS builder

# go-coap/v3 requiere Go >= 1.24; permitimos descarga automática del toolchain
ENV GOTOOLCHAIN=auto

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o nodo ./cmd/nodo

# Imagen final ligera
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/nodo .

EXPOSE 5683/udp
CMD ["./nodo"]
