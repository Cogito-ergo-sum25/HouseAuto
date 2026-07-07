# Etapa de compilación - Actualizado a Go 1.24
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Inicializar y descargar dependencias
COPY go.mod ./
RUN go mod tidy

# Copiar el código fuente
COPY . .

# Compilar desde la nueva ubicación
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Etapa final de producción
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY frontend/ ./frontend/
EXPOSE 8080
CMD ["./main"]