update_frontend_client() {
  rm -rf ../frontend/neosync-api-client
  mkdir -p ../frontend/neosync-api-client
  mv gen/es/protos/** ../frontend/neosync-api-client
  rm -rf gen/es
}

update_docs() {
  rm -rf ../docs/protos
  mkdir -p ../docs/protos
  mv gen/docs/** ../docs/protos
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
  --volume "./buf.work.yaml:/workspace/buf.work.yaml" \
  --volume "./buf.gen.yaml:/workspace/buf.gen.yaml" \
  --volume "./protos:/workspace/protos" \
  --workdir "/workspace" \
  "bufbuild/buf:${BUF_VERSION}" generate &

# sqlc
docker run --rm -i \
  --volume "./gen:/workspace/gen" \
  --volume "./sql:/workspace/sql" \
  --volume "./sqlc.yaml:/workspace/sqlc.yaml" \
  --workdir "/workspace" \
  "sqlc/sqlc:${SQLC_VERSION}" generate &
wait

update_frontend_client &
update_docs &

wait
