#!/bin/sh

if [ ! -z "${TILT_HOST}" ]; then
  # when invoked by tilt set the OS so that binary will run on linux container
  CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/mgmt cmd/mgmt/*.go
  #CGO_ENABLED=0 GOOS=linux go build -gcflags="all=-N -l" -o bin/mgmt cmd/mgmt/*.go
else
  CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/mgmt cmd/mgmt/*.go
  #CGO_ENABLED=0 $(GO) build -gcflags="all=-N -l" -o bin/mgmt cmd/mgmt/*.go
fi
