PROTO_DIR := proto
GEN_DIR := gen
MODULE := dos 
PROJECT := dos

PROTO_FILES := $(shell find $(PROTO_DIR) -name "*.proto")
COMPOSE = docker compose -f docker-compose.yml -p $(PROJECT)

.PHONY: gen build up down client build-e2e e2e

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
	docker build -f deploy/docker/client.Dockerfile -t dos-client:latest .

build-e2e:
	docker build -f deploy/docker/e2e.Dockerfile -t dos-e2e:latest .

up:
	$(COMPOSE) --profile main up

down:
	$(COMPOSE) --profile main down -v --remove-orphans

client:
	$(COMPOSE) run --rm client /bin/bash

e2e:
	$(COMPOSE) --profile e2e run --rm e2e
