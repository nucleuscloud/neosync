#!/bin/sh

docker run --rm \
  --volume "./charts/neosync:/helm-docs/charts/neosync" \
  --volume "./backend/charts:/helm-docs/backend/charts" \
  --volume "./worker/charts:/helm-docs/worker/charts" \
  --volume "./frontend/apps/web/charts:/helm-docs/frontend/apps/web/charts" \
  -u "$(id -u)" \
  jnorwood/helm-docs:v1.13.1 --document-dependency-values=true
