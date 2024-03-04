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
	cd ./backend && make all
.PHONY: build-backend

build-worker:
	cd ./worker && make all
.PHONY: build-worker

build-cli:
	cd ./cli && make all
.PHONY: build-cli

clean: clean-backend clean-worker clean-cli
.PHONY: clean

clean-backend:
	cd ./backend && make clean
.PHONY: clean-backend

clean-worker:
	cd ./worker && make clean
.PHONY: clean-worker

clean-cli:
	cd ./cli && make clean
.PHONY: clean-cli

rebuild: clean build
.PHONY: rebuild

compose-up:
	docker compose -f $(TEMPORAL_COMPOSE_FILE) up -d
	docker compose -f $(PROD_COMPOSE_FILE) pull
	docker compose -f $(PROD_COMPOSE_FILE) up -d
.PHONY: compose-up

compose-down:
	docker compose -f $(TEMPORAL_COMPOSE_FILE) down
	docker compose -f $(PROD_COMPOSE_FILE) down
.PHONY: compose-down

compose-auth-up:
	docker compose -f $(TEMPORAL_COMPOSE_FILE) up -d
	docker compose -f $(PROD_AUTH_COMPOSE_FILE) pull
	docker compose -f $(PROD_AUTH_COMPOSE_FILE) up -d
.PHONY: compose-auth-up

compose-auth-down:
	docker compose -f $(TEMPORAL_COMPOSE_FILE) down
	docker compose -f $(PROD_AUTH_COMPOSE_FILE) down
.PHONY: compose-auth-down

compose-dev-up: rebuild
	docker compose -f $(TEMPORAL_COMPOSE_FILE) up -d
	docker compose -f $(DEV_COMPOSE_FILE) watch
.PHONY: compose-dev-up

compose-dev-down: rebuild
	docker compose -f $(TEMPORAL_COMPOSE_FILE) down
	docker compose -f $(DEV_COMPOSE_FILE) down
.PHONY: compose-dev-down

compose-dev-auth-up: rebuild
	docker compose -f $(TEMPORAL_COMPOSE_FILE) up -d
	docker compose -f $(DEV_AUTH_COMPOSE_FILE) watch
.PHONY: compose-dev-auth-up

compose-dev-auth-down: rebuild
	docker compose -f $(TEMPORAL_COMPOSE_FILE) down
	docker compose -f $(DEV_AUTH_COMPOSE_FILE) down
.PHONY: compose-dev-auth-down

goworksync:
	go work sync
	cd worker && go mod tidy
	cd cli && go mod tidy
	cd backend && go mod tidy
.PHONY: goworksync
