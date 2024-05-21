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

# `buf format -w` writes to each file, so causes tilt to loop. instead we instruct buf
# to just output the changes in diff format and apply those.
docker run --rm -i \
  --volume "./protos:/workspace/protos" \
  --workdir "/workspace" \
  "bufbuild/buf:${BUF_VERSION}" format -d | patch --quiet

# buf generate
docker run --rm -i \
  --volume "./gen:/workspace/gen" \
  --volume "./buf.yaml:/workspace/buf.yaml" \
  --volume "./buf.lock:/workspace/buf.lock" \
  --volume "./buf.gen.yaml:/workspace/buf.gen.yaml" \
  --volume "./protos:/workspace/protos" \
  --workdir "/workspace" \
  "bufbuild/buf:${BUF_VERSION}" generate &

# sqlc
docker run --rm -i \
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
