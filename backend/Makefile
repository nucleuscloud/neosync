.PHONY: help default \
				vet \
				gen build dbuild \
				run \
				clean \
				lint \
				install-tools install-go-tools

default: help

help:
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_\/]+:.*##/ { printf "\033[36m%-30s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

vet: ## Runs go vet
	go vet ./...

build: gen vet ## Generates code, runs go vet, and builds the project
	scripts/build.sh

dbuild: build

run: ## Runs the project
	./bin/mgmt serve connect

gen: ## Runs the gen script
	@cd .. && ./scripts/backend-gen.sh

clean: ## Cleans the bin directory
	rm -rf bin

lint: ## Runs the Go Linter
	golangci-lint run

install-tools: install-go-tools ## Installs necessary tools

install-go-tools: ## Installs Go tools with postgres tag
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate
