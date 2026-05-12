PROTO_DIR := proto
GEN_DIR := gen
MODULE := dos 
BIN_DIR := bin
PROJECT_NAME := dos

PROTO_FILES := $(shell find $(PROTO_DIR) -name "*.proto")

.PHONY: gen build up down down-clean clean build-client

gen:
	protoc -I $(PROTO_DIR) \
		--go_out=. \
		--go_opt=module=$(MODULE) \
		--go-grpc_out=. \
		--go-grpc_opt=module=$(MODULE) \
		$(PROTO_FILES)

build:
	docker build -f deploy/docker/master.Dockerfile -t dos-master:latest .
	docker build -f deploy/docker/storage.Dockerfile -t dos-storage:latest .

up:
	docker compose up

down:
	docker compose down

down-clean:
	docker compose down -v

clean:
	rm -rf $(BIN_DIR)

build-client:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(PROJECT_NAME) ./cmd/client/cli/
	cp cmd/client/cli/config.yml $(BIN_DIR)/config.yml
