#!/bin/sh

DBPORT="${DB_PORT:-5432}"

docker run -e POSTGRES_PASSWORD=foofar -e POSTGRES_DB=nucleus -p $DBPORT:5432 postgres:15
