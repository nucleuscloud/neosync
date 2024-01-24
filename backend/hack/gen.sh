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

update_python_client() {
  rm -rf ../ml/backend/mgmt
  rm -rf ../ml/backend/buf
  mv gen/python/protos** ../ml/backend
  rm -rf gen/python
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
  --volume "./buf.work.yaml:/workspace/buf.work.yaml" \
  --volume "./buf.gen.yaml:/workspace/buf.gen.yaml" \
  --volume "./protos:/workspace/protos" \
  --workdir "/workspace" \
  "bufbuild/buf:${BUF_VERSION}" generate &

  # buf generate
docker run --rm -i \
  --volume "./gen:/workspace/gen" \
  --volume "./buf.work.yaml:/workspace/buf.work.yaml" \
  --volume "./buf.gen.python.yaml:/workspace/buf.gen.yaml" \
  --volume "./protos:/workspace/protos" \
  --workdir "/workspace" \
  "bufbuild/buf:${BUF_VERSION}" generate --include-imports &

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
update_python_client &

wait
