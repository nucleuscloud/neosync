GO=go

default: all

all: build

vet:
	$(GO) vet ./...

build: gen vet
	hack/build.sh
.PHONY: build

run:
	./bin/mgmt serve connect
.PHONY: run

# build-debug:
# 	CGO_ENABLED=0 $(GO) build -gcflags="all=-N -l" -o bin/mgmt cmd/mgmt/*.go
# .PHONY: build-debug

protos: gen
.PHONY: protos

gen:
	./hack/gen.sh
.PHONY: gen

install-tools: install-go-tools
.PHONY: install-tools

install-go-tools:
	$(GO) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate
.PHONY: install-go-tools

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[/.a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
.PHONY: help
