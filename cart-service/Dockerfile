FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o cart-service .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/cart-service .

EXPOSE 8003
CMD ["./cart-service"]