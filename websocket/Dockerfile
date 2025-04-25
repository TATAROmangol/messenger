
FROM golang:1.24-alpine3.21 AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./websocket ./main.go

FROM alpine:3.21
WORKDIR /app

COPY --from=builder /app/websocket /

EXPOSE 3000

CMD ["/websocket"]