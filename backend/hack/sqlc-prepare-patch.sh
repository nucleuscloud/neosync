#!/bin/sh

databaseUri=$1
FILEPATH=$2

if [ "$FILEPATH" == "" ]; then
  FILEPATH="sqlc.yaml"
fi

yq -i ".sql[0].database.uri = \"$databaseUri\"" "$FILEPATH"

# creates array if rules doesn't exist, concats rules + the sqlc/db-prepare
# shouldn't need to de-dupe here because by default we don't have sqlc/db-prepare in the standard config
yq -i ".sql[0].rules = [] + .sql[0].rules + \"sqlc/db-prepare\"" "$FILEPATH"
