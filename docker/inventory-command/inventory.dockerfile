FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o inventory-service ./cmd/inventory-command

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/inventory-service .
EXPOSE 8080

CMD ["./inventory-service"]