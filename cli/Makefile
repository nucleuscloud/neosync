.PHONY: help default \
				vet \
				build dbuild \
				clean \
				lint \

default: help

help:
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_\/]+:.*##/ { printf "\033[36m%-30s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)


BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT ?= $(shell git rev-parse HEAD)
VERSION ?=$(shell echo $(shell git describe --tags --abbrev=0 2> /dev/null || echo "untagged" ) | tr -d '\n')


vet:
	go vet ./...

build: vet ## Runs go vet and builds the project
	# Note the ldflags here are used for local builds only. To see the ldflags at release time, check the .goreleaser.yaml file
	CGO_ENABLED=0 go build -o bin/neosync -ldflags="-s -w -X 'github.com/nucleuscloud/neosync/cli/internal/version.buildDate=${BUILD_DATE}' -X 'github.com/nucleuscloud/neosync/cli/internal/version.gitCommit=${GIT_COMMIT}' -X 'github.com/nucleuscloud/neosync/cli/internal/version.gitVersion=${VERSION}'" cmd/neosync/*.go

dbuild: vet ## Runs go vet and builds thd project specifically for Linux
# Note the ldflags here are used for local builds only. To see the ldflags at release time, check the .goreleaser.yaml file
	CGO_ENABLED=0 GOOS=linux go build -o bin/neosync -ldflags="-s -w -X 'github.com/nucleuscloud/neosync/cli/internal/version.buildDate=${BUILD_DATE}' -X 'github.com/nucleuscloud/neosync/cli/internal/version.gitCommit=${GIT_COMMIT}' -X 'github.com/nucleuscloud/neosync/cli/internal/version.gitVersion=${VERSION}'" cmd/neosync/*.go

clean: ## Cleans the bin directory
	rm -rf bin

lint:
	golangci-lint run
