PROTO_DIR := proto
GEN_DIR := gen
MODULE := dos 
PROJECT := dos

PROTO_FILES := $(shell find $(PROTO_DIR) -name "*.proto")
COMPOSE = docker compose -f docker-compose.yml -p $(PROJECT)

.PHONY: gen build up down client build-e2e e2e load wait 
.PONNY: logs master-logs storage-logs test-logs

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

build-test:
	docker build -f deploy/docker/test.Dockerfile -t dos-test:latest .

up:
	$(COMPOSE) --profile main up -d

down:
	$(COMPOSE) --profile main down -v --remove-orphans

client:
	$(COMPOSE) run --rm client /bin/bash

e2e:
	$(COMPOSE) --profile test run --rm e2e-test

load:
	$(COMPOSE) --profile test run --rm load-test /bin/bash

wait:
	$(COMPOSE) --profile test run --rm e2e-test python tests/support/wait_cluster.py

logs: 
	$(COMPOSE) --profile main logs -f --no-color --tail=300

master-logs: 
	$(COMPOSE) --profile master logs --no-color --tail=300

storage-logs:
	$(COMPOSE) --profile storage logs --no-color --tail=300

test-logs:
	$(COMPOSE) --profile test logs --no-color --tail=300
