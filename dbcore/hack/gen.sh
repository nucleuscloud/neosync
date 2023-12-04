#!/bin/sh

SQLC_VERSION=$(cat SQLC_VERSION)

docker run --rm -i --volume "./:/src" --workdir "/src" "sqlc/sqlc:${SQLC_VERSION}" generate &
wait
