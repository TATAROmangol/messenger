
FROM golang:1.24-alpine3.21 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./chat ./cmd/chat/main.go

FROM alpine:3.21
WORKDIR /app

COPY --from=builder /app/.env ./
COPY --from=builder /app/chat ./cmd/
COPY --from=builder /app/migrations/* ./migrations/

EXPOSE 8080

CMD ["./cmd/chat"]