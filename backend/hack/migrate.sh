#!/bin/sh
#
# To customize for non-dev environments, set the following environment variables with
# their appropriate values.
#
#   PG_USERNAME
#   PG_PASSWORD
#   PG_HOSTNAME
#
# Also! Set SKIP_SCHEMA_CREATION=true to skip schema bootstrapping that migrate needs
# b/c we do not use the public schema.
#
# Note that this script did not bootstrap initial stage/production dbs. That
# code is in terraform.
#

#DEBUG=1
debug() {
    if [ -n "$DEBUG" ]; then
        echo "DEBUG: $1"
    fi
}

# odd quoting for migrate, which will attempt to create this table in public schema
# if you don't quote it with double quotes like this.
MIGRATION_TABLE_NAME='"public"."neosync_api_schema_migrations"'

PG_LOGIN="${PG_USERNAME:-postgres}:${PG_PASSWORD:-foofar}"
PG_HOSTNAME=${PG_HOSTNAME:-"localhost:5433/neosync"}
PG_OPTIONS="x-migrations-table=${MIGRATION_TABLE_NAME}&x-migrations-table-quoted=true"

if [ "$PG_HOSTNAME" = "localhost:5433/neosync" ]; then
    PG_OPTIONS="${PG_OPTIONS}&sslmode=disable"
fi
if [ "$PG_HOSTNAME" = "postgresql:5433/neosync" ]; then
    PG_OPTIONS="${PG_OPTIONS}&sslmode=disable"
fi

SCRIPT_DIR=$(basename "$0")

if [ -z "$1" ]; then
    echo "Must specify up or down; you can optionally pass in number of steps as well, i.e., 'migrate up 1'"
    exit 1
fi
cmd="$1"

if ! command -v migrate > /dev/null ; then
    echo "Must have golang-migrate installed. Run 'make install' if OSX or consult your package manager."
    exit 1
fi

PG_CONNECT_STR="postgres://${PG_LOGIN}@${PG_HOSTNAME}"

debug "PG_USERNAME: ${PG_USERNAME}"
debug "PG_LOGIN:    ${PG_LOGIN}"
debug "PG_OPTIONS:  ${PG_OPTIONS}"
debug "PG_HOSTNAME: ${PG_HOSTNAME}"


# only need schema creation for local dev, so allow skipping by setting this env var
if [ -z "${SKIP_SCHEMA_CREATION}" ] || [ "${SKIP_SCHEMA_CREATION}" = "false" ] && [ "${cmd}" = "up" ]; then
    if ! command -v psql > /dev/null ; then
        echo "Must have psql installed. Check your package manager."
        exit 1
    fi

    debug "Attempting to create neosync_api schema"

    # shellcheck disable=2086 # need expansion here
    if ! psql ${PG_CONNECT_STR} -c 'CREATE SCHEMA IF NOT EXISTS neosync_api'; then
        echo "error upserting neosync_api schema"
        exit 1
    fi
fi

PG_CONNECT_STR="${PG_CONNECT_STR}?${PG_OPTIONS}"

debug "PG_CONNECT_STR: ${PG_CONNECT_STR}"

# make stderr go through stdout so it doesn't end up an err in datadog
# shellcheck disable=2086
migrate -path "${SCRIPT_DIR}/../migrations" -database "${PG_CONNECT_STR}" "$@" 2>&1
