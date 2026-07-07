# Etapa de compilación - Actualizado a Go 1.24
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copiar todo el código fuente
COPY . .

# Generar go.sum y descargar dependencias basándose en el código copiado
RUN go mod tidy

# Compilar desde la nueva ubicación
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Etapa final de producción
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY frontend/ ./frontend/
EXPOSE 8080
CMD ["./main"]