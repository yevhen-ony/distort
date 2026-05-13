PROTO_DIR := proto
GEN_DIR := gen
MODULE := dos 
PROJECT := dos

PROTO_FILES := $(shell find $(PROTO_DIR) -name "*.proto")
COMPOSE = docker compose -f docker-compose.yml -p $(PROJECT)

.PHONY: gen build up down down-clean clean build-client

gen:
	protoc -I $(PROTO_DIR) \
		--go_out=. \
		--go_opt=module=$(MODULE) \
		--go-grpc_out=. \
		--go-grpc_opt=module=$(MODULE) \
		$(PROTO_FILES)

build:
	docker build --no-cache -f deploy/docker/master.Dockerfile -t dos-master:latest .
	docker build --no-cache -f deploy/docker/storage.Dockerfile -t dos-storage:latest .
	docker build --no-cache -f deploy/docker/client.Dockerfile -t dos-client:latest .

up:
	$(COMPOSE) --profile main up

down:
	$(COMPOSE) --profile main down --remove-orphans

down-clean:
	$(COMPOSE) --profile main down -v

client:
	docker compose run --rm client /bin/bash
