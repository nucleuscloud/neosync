#!/bin/sh

FILE=$(readlink -f "$0")
NEOSYNC_ROOT="$(dirname "$FILE")/../../"

ctlptl_create()
{
  if ! command -v ctlptl > /dev/null ; then
    echo "requires ctptl to run, see https://github.com/tilt-dev/ctlptl"
    exit 1
  fi


  NEOSYNC_DEV_HOSTPATH="${NEOSYNC_ROOT}/.data"
  mkdir -p "$NEOSYNC_DEV_HOSTPATH"
  chmod 777 "$NEOSYNC_DEV_HOSTPATH"
  sed 's|{NEOSYNC_DEV_HOSTPATH}|'"$NEOSYNC_DEV_HOSTPATH"'|' < "$NEOSYNC_ROOT/tilt/kind/cluster.yaml" | ctlptl delete -f -
}

ctlptl_create
