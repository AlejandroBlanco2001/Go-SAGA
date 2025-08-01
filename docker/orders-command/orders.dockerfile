FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o orders-service ./cmd/orders-command

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/orders-service .
EXPOSE 8080

CMD ["./orders-service"]