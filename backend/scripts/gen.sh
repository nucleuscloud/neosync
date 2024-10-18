update_frontend_client() {
  rm -rf ../frontend/packages/sdk/src/client/mgmt
  mv gen/es/protos/** ../frontend/packages/sdk/src/client
  rm -rf gen/es
}

update_docs() {
  rm -rf ../docs/protos/data
  mkdir -p ../docs/protos/data
  mv gen/docs/** ../docs/protos/data
  rm -rf gen/docs
}

BUF_VERSION=$(cat BUF_VERSION)
SQLC_VERSION=$(cat SQLC_VERSION)

# Get the current user ID and group ID
USER_ID=$(id -u)
GROUP_ID=$(id -g)

# Determine cache directory based on XDG_CACHE_HOME if set, otherwise use $HOME/.cache/buf
if [ -z "${XDG_CACHE_HOME}" ]; then
  BUF_CACHE_DIRECTORY="${HOME}/.cache/buf"
else
  BUF_CACHE_DIRECTORY="${XDG_CACHE_HOME}/buf"
fi
mkdir -p ${BUF_CACHE_DIRECTORY}


# `buf format -w` writes to each file, so causes tilt to loop. instead we instruct buf
# to just output the changes in diff format and apply those.
docker run --rm -i \
  --user "${USER_ID}:${GROUP_ID}" \
  --env BUF_CACHE_DIR=/workspace/.cache \
  --volume "./protos:/protos" \
  --volume "${BUF_CACHE_DIRECTORY}:/workspace/.cache" \
  --workdir "/protos" \
  "bufbuild/buf:${BUF_VERSION}" format -d | patch -d ./protos -p0 --quiet

# Detect host architecture
HOST_ARCH=$(uname -m)
if [ "$HOST_ARCH" = "aarch64" ] || [ "$HOST_ARCH" = "arm64" ]; then
    PLUGIN_BINARY="protoc-gen-connect-openapi_0.13.0_linux_arm64"
elif [ "$HOST_ARCH" = "x86_64" ]; then
    PLUGIN_BINARY="protoc-gen-connect-openapi_0.13.0_linux_amd64"
else
    echo "Unsupported architecture: $HOST_ARCH"
    exit 1
fi

ENV_FILE="./.env.dev.secrets"
BUF_GENERATE_CMD="docker run --rm -i"
# Check if the environment file exists and include it if it does
if [ -f "$ENV_FILE" ]; then
  BUF_GENERATE_CMD="$BUF_GENERATE_CMD --env-file $ENV_FILE"
fi

BUF_GENERATE_CMD="$BUF_GENERATE_CMD \
  --user \"${USER_ID}:${GROUP_ID}\" \
  --env BUF_CACHE_DIR=/workspace/.cache \
  --volume \"./gen:/workspace/gen\" \
  --volume \"./buf.yaml:/workspace/buf.yaml\" \
  --volume \"./buf.lock:/workspace/buf.lock\" \
  --volume \"./buf.gen.yaml:/workspace/buf.gen.yaml\" \
  --volume \"./protos:/workspace/protos\" \
  --volume \"${BUF_CACHE_DIRECTORY}:/workspace/.cache\" \
  --volume \"./scripts/protoc-gen-connect-openapi/${PLUGIN_BINARY}:/usr/local/bin/protoc-gen-connect-openapi\" \
  --workdir \"/workspace\" \
  \"bufbuild/buf:${BUF_VERSION}\" generate &"

eval $BUF_GENERATE_CMD

# sqlc
docker run --rm -i \
  --user "${USER_ID}:${GROUP_ID}" \
  --volume "./gen:/workspace/gen" \
  --volume "./sql:/workspace/sql" \
  --volume "./pkg/dbschemas/sql:/workspace/pkg/dbschemas/sql" \
  --volume "./sqlc.yaml:/workspace/sqlc.yaml" \
  --workdir "/workspace" \
  "sqlc/sqlc:${SQLC_VERSION}" generate &
wait

update_frontend_client &
update_docs &

wait
