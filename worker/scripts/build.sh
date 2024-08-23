#!/bin/sh

if [ ! -z "${TILT_HOST}" ]; then
  # when invoked by tilt set the OS so that binary will run on linux container
  GOOS=linux go build -ldflags="-s -w" -o bin/worker cmd/worker/*.go
elif [ ! -z "${GOOS}" ]; then
  GOOS="${GOOS}" go build -ldflags="-s -w" -o bin/worker cmd/worker/*.go
else
  go build -ldflags="-s -w" -o bin/worker cmd/worker/*.go
fi
