GO=go

default: all

all: build

vet:
	$(GO) vet ./...

lint:
	golangci-lint run
.PHONY: lint

build: vet
	# Note the ldflags here are used for local builds only. To see the ldflags at release time, check the .goreleaser.yaml file
	# $(GO) build -o bin/nucleus -ldflags="-X 'github.com/nucleuscloud/cli/internal/version.buildDate=${BUILD_DATE}' -X 'github.com/nucleuscloud/cli/internal/version.gitCommit=${GIT_COMMIT}' -X 'github.com/nucleuscloud/cli/internal/version.gitVersion=${VERSION}'"
	CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o bin/neosync cmd/neosync/*.go
.PHONY: build



help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[/.a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
.PHONY: help
