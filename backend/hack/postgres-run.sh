#!/bin/sh

docker run -e POSTGRES_PASSWORD=foofar -e POSTGRES_DB=nucleus -p 5432:5432 postgres:14.4
