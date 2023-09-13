#!/usr/bin/env bash

if ! command -v migrate > /dev/null; then
    echo "Could not find migrate command in your path; try running 'make install-tools'"
    exit 1
fi

if [ -z "$1" ]; then
    echo "Must specify the migration name, like users-add-expiration-date"
    exit 1
fi

SCRIPT_DIR=$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)

# use epoch time so unlikely to clash
set -v
migrate create -ext sql -dir "${SCRIPT_DIR}/../migrations" "$1"
