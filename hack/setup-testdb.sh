#!/bin/sh

helmfile sync -f hack/testdbs/helmfile.yaml

SRC_PORT="5433"
DST_PORT="5434"

pkill -f "kubectl port-forward -n default svc/postgresql1 ${SRC_PORT}:5432"
pkill -f "kubectl port-forward -n default svc/postgresql2 ${DST_PORT}:5432"

kubectl port-forward -n default svc/postgresql1 ${SRC_PORT}:5432 &
kubectl port-forward -n default svc/postgresql2 ${DST_PORT}:5432 &

sleep 5

SRC_CONN="postgresql://postgres:foofar@localhost:${SRC_PORT}/nucleus?sslmode=disable"
DST_CONN="postgresql://postgres:foofar@localhost:${DST_PORT}/nucleus?sslmode=disable"

psql '${SRC_CONN}' -c 'CREATE SCHEMA IF NOT EXISTS neosync_api'
psql '${DST_CONN}' -c 'CREATE SCHEMA IF NOT EXISTS neosync_api'

migrate -path backend/migrations -database "${SRC_CONN}" up
migrate -path backend/migrations -database "${DST_CONN}" up

# seed data into source db
psql "${SRC_CONN}" -f hack/source-data.sql

pkill -f "kubectl port-forward -n default svc/postgresql1 ${SRC_PORT}:5432"
pkill -f "kubectl port-forward -n default svc/postgresql2 ${DST_PORT}:5432"
