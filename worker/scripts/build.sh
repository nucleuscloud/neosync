#!/bin/sh

if [ ! -z "${TILT_HOST}" ]; then
  # when invoked by tilt set the OS so that binary will run on linux container
  CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/worker cmd/worker/*.go
elif [ ! -z "${GOOS}" ]; then
  CGO_ENABLED=0 GOOS="${GOOS}" go build -ldflags="-s -w" -o bin/worker cmd/worker/*.go
else
  CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/worker cmd/worker/*.go
fi
