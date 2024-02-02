default: help

TEMPORAL_COMPOSE_FILE = temporal/compose.yml
PROD_COMPOSE_FILE = compose/compose-prod.yml
PROD_AUTH_COMPOSE_FILE = compose/compose-auth-prod.yml
DEV_COMPOSE_FILE = compose.yml
DEV_AUTH_COMPOSE_FILE = compose/compose-auth.yml

##@
cluster-create:
	sh ./tilt/scripts/cluster-create.sh
.PHONY: cluster-create

cluster-destroy:
	bash ./tilt/scripts/assert-context.sh
	sh ./tilt/scripts/cluster-destroy.sh
.PHONY: cluster-destroy

build: build-backend build-worker build-cli
.PHONY: build

build-backend:
	sh -c "cd ./backend && make all"
.PHONY: build-backend

build-worker:
	sh -c "cd ./worker && make all"
.PHONY: build-worker

build-cli:
	sh -c "cd ./cli && make all"
.PHONY: build-cli

clean: clean-backend clean-worker clean-cli
.PHONY: clean

clean-backend:
	sh -c "cd ./backend && make clean"
.PHONY: clean-backend

clean-worker:
	sh -c "cd ./worker && make clean"
.PHONY: clean-worker

clean-cli:
	sh -c "cd ./cli && make clean"
.PHONY: clean-cli

rebuild: clean build
.PHONY: rebuild

compose-up:
	sh -c "docker compose -f $(TEMPORAL_COMPOSE_FILE) up -d"
	sh -c "docker compose -f $(PROD_COMPOSE_FILE) up -d"
.PHONY: compose-up

compose-down:
	sh -c "docker compose -f $(TEMPORAL_COMPOSE_FILE) down"
	sh -c "docker compose -f $(PROD_COMPOSE_FILE) down"
.PHONY: compose-down

compose-auth-up:
	sh -c "docker compose -f $(TEMPORAL_COMPOSE_FILE) up -d"
	sh -c "docker compose -f $(PROD_AUTH_COMPOSE_FILE) up -d"
.PHONY: compose-auth-up

compose-auth-down:
	sh -c "docker compose -f $(TEMPORAL_COMPOSE_FILE) down"
	sh -c "docker compose -f $(PROD_AUTH_COMPOSE_FILE) down"
.PHONY: compose-auth-down

compose-dev-up: rebuild
	sh -c "docker compose -f $(TEMPORAL_COMPOSE_FILE) up -d"
	sh -c "docker compose -f $(DEV_COMPOSE_FILE) watch"
.PHONY: compose-dev-up

compose-dev-down: rebuild
	sh -c "docker compose -f $(TEMPORAL_COMPOSE_FILE) down"
	sh -c "docker compose -f $(DEV_COMPOSE_FILE) down"
.PHONY: compose-dev-down

compose-dev-auth-up: rebuild
	sh -c "docker compose -f $(TEMPORAL_COMPOSE_FILE) up -d"
	sh -c "docker compose -f $(DEV_AUTH_COMPOSE_FILE) watch"
.PHONY: compose-dev-auth-up

compose-dev-auth-down: rebuild
	sh -c "docker compose -f $(TEMPORAL_COMPOSE_FILE) down"
	sh -c "docker compose -f $(DEV_AUTH_COMPOSE_FILE) down"
.PHONY: compose-dev-auth-down

goworksync:
	go work sync
	cd worker && go mod tidy
	cd cli && go mod tidy
	cd backend && go mod tidy
	cd terraform-provider && go mod tidy
.PHONY: goworksync
