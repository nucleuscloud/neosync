.PHONY: help default \
        cluster/create cluster/destroy \
        build build/backend build/worker build/cli \
				install/frontend \
        clean clean/backend clean/worker clean/cli \
        compose/up compose/down \
        compose/auth/up compose/auth/down \
        compose/dev/up compose/dev/down \
        compose/dev/auth/up compose/dev/auth/down
default: help

help:
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_\/]+:.*##/ { printf "\033[36m%-30s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

PROD_COMPOSE_FILE = compose.yml
PROD_AUTH_COMPOSE_FILE = compose.auth.yml
DEV_COMPOSE_FILE = compose.dev.yml
DEV_AUTH_COMPOSE_FILE = compose.auth.dev.yml

# Cluster Management
cluster/create: ## Creates a local K8s Cluster
	sh ./tilt/scripts/cluster-create.sh

cluster/destroy: ## Destroys a local K8s Cluster
	bash ./tilt/scripts/assert-context.sh
	sh ./tilt/scripts/cluster-destroy.sh

# Building
build: build/backend build/worker build/cli install/frontend ## Builds the project

build/backend: ## Builds the backend
	cd ./backend && make all

build/worker: ## Builds the worker
	cd ./worker && make all

build/cli: ## Builds the CLI
	cd ./cli && make all

install/frontend: ## Runs npm install for the frontend
	cd ./frontend && npm install

# Cleaning
clean: clean/backend clean/worker clean/cli ## Cleans the project

clean/backend: ## Cleans the backend
	cd ./backend && make clean

clean/worker: ## Cleans the worker
	cd ./worker && make clean

clean/cli: ## Cleans the CLI
	cd ./cli && make clean

# Compose Management
compose/up: ## Composes up the production environment
	docker compose -f $(PROD_COMPOSE_FILE) up -d

compose/down: ## Composes down the production environment
	docker compose -f $(PROD_COMPOSE_FILE) down

compose/auth/up: ## Composes up the production environment with auth
	docker compose -f $(PROD_COMPOSE_FILE) -f $(PROD_AUTH_COMPOSE_FILE) up -d

compose/auth/down: ## Composes down the production environment with auth
	docker compose -f $(PROD_COMPOSE_FILE) -f $(PROD_AUTH_COMPOSE_FILE) down

compose/dev/up: ## Composes up the development environment
	docker compose -f $(DEV_COMPOSE_FILE) watch

compose/dev/down: ## Composes down the development environment
	docker compose -f $(DEV_COMPOSE_FILE) down

compose/dev/auth/up: ## Composes up the development environment with auth
	docker compose -f $(DEV_COMPOSE_FILE) -f $(DEV_AUTH_COMPOSE_FILE) watch

compose/dev/auth/down: ## Composes down the development environment with auth
	docker compose -f $(DEV_COMPOSE_FILE) -f $(DEV_AUTH_COMPOSE_FILE) down
