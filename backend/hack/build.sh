#!/bin/sh

BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse HEAD)
VERSION=$(git describe --tags --abbrev=0 | tr -d '\n')

if [ ! -z "${TILT_HOST}" ]; then
  # when invoked by tilt set the OS so that binary will run on linux container
  # Note the ldflags set here are used only for local builds.
  CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X 'github.com/nucleuscloud/neosync/backend/internal/version.gitVersion=$VERSION' -X 'github.com/nucleuscloud/neosync/backend/internal/version.buildDate=$BUILD_DATE' -X 'github.com/nucleuscloud/neosync/backend/internal/version.gitCommit=$GIT_COMMIT'" -o bin/mgmt cmd/mgmt/*.go
  #CGO_ENABLED=0 GOOS=linux go build -gcflags="all=-N -l" -o bin/mgmt cmd/mgmt/*.go
elif [ ! -z "${GOOS}" ]; then
  CGO_ENABLED=0 GOOS="${GOOS}" go build -ldflags="-s -w -X 'github.com/nucleuscloud/neosync/backend/internal/version.gitVersion=$VERSION' -X 'github.com/nucleuscloud/neosync/backend/internal/version.buildDate=$BUILD_DATE' -X 'github.com/nucleuscloud/neosync/backend/internal/version.gitCommit=$GIT_COMMIT'" -o bin/mgmt cmd/mgmt/*.go
else
  CGO_ENABLED=0 go build -ldflags="-s -w -X 'github.com/nucleuscloud/neosync/backend/internal/version.gitVersion=$VERSION' -X 'github.com/nucleuscloud/neosync/backend/internal/version.buildDate=$BUILD_DATE' -X 'github.com/nucleuscloud/neosync/backend/internal/version.gitCommit=$GIT_COMMIT'" -o bin/mgmt cmd/mgmt/*.go
  #CGO_ENABLED=0 $(GO) build -gcflags="all=-N -l" -o bin/mgmt cmd/mgmt/*.go
fi
