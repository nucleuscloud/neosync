update_frontend_client() {
  rm -rf ../frontend/neosync-api-client
  mkdir -p ../frontend/neosync-api-client
  mv gen/es/protos/** ../frontend/neosync-api-client
  rm -rf gen/es
}

# `buf format -w` writes to each file, so causes tilt to loop. instead we instruct buf
# to just output the changes in diff format and apply those.
BUF_VERSION=$(cat BUF_VERSION)
SQLC_VERSION=$(cat SQLC_VERSION)

# buf fmt
docker run --rm -i \
  --volume "./protos:/workspace" \
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
docker run --rm -i \
  --volume "./gen:/workspace/gen" \
  --volume "./buf.work.yaml:/workspace/buf.work.yaml" \
  --volume "./buf-es.gen.yaml:/workspace/buf.gen.yaml" \
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

# docker run --rm -i --volume "./:/src" --workdir "/src" "vektra/mockery:${MOCKERY_VERSION}" &
update_frontend_client &

wait
