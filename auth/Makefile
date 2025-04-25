SOURCES = ./cmd/* ./internal/config/* ./internal/transport/* ./pkg/logger/* ./pkg/postgres/* ./go.*

all:

pb-gen:
	protoc \
	--go_out=./pkg/ --go_opt=paths=source_relative \
	--go-grpc_out=./pkg/ --go-grpc_opt=paths=source_relative ./api/auth.proto

build: $(SOURCES)
	go build -o ./build/auth ./cmd/auth.go

run:
	./build/auth

docker-up:
	sudo docker compose up

docker-rebuild:
	sudo docker compose up --build

pg-up:
	sudo docker compose -f ./pg-compose.yml up

test:
	go test ./... -cover