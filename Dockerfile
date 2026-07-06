# Etapa de compilación
FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN go mod init porton-web && go get github.com/eclipse/paho.mqtt.golang
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Etapa final de producción
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY index.html .
EXPOSE 8080
CMD ["./main"]