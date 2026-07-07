# Etapa 1: Compilación
FROM golang:alpine AS builder
WORKDIR /app

# Descargar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código y compilar el binario
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Etapa 2: Producción (Imagen súper ligera)
FROM alpine:latest
WORKDIR /app

# Instalar dependencias necesarias para SQLite
RUN apk --no-cache add ca-certificates tzdata

# Copiar el binario compilado
COPY --from=builder /app/main .
# Copiar el frontend
COPY --from=builder /app/frontend/public ./frontend/public

EXPOSE 8080

# Iniciar la aplicación
CMD ["./main"]