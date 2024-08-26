.PHONY: help default \
				vet \
				gen build dbuild \
				run \
				clean \
				lint \

default: help

help:
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_\/]+:.*##/ { printf "\033[36m%-30s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

vet: ## Runs go vet
	go vet ./...

build: gen vet ## Generates code, runs go vet and builds the project
	scripts/build.sh

dbuild: build

run: ## Runs the project
	./bin/sync serve

gen: ## Runs go generate
	go generate ./...

clean: ## Cleans the bin directory
	rm -rf bin

lint: ## Runs the Go Linter
	golangic-lint run
