#!/bin/sh
#
# Could this not be a heredoc in the dockerfile?
# todo: we should remove this and just install with go install tools

PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:'])
MACHINE=$(uname -m)
if [ "$MACHINE" = "x86_64" ]; then
    MACHINE="amd64"
elif [ "$MACHINE" = "aarch64" ]; then
  MACHINE=arm64
fi

curl -sSL "https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.${PLATFORM}-${MACHINE}.tar.gz" \
  -o migrate.tar.gz && \
  tar xzvf migrate.tar.gz migrate -C $HOME/.local/bin && \
  rm  migrate.tar.gz
