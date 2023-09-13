# `buf format -w` writes to each file, so causes tilt to loop. instead we instruct buf
# to just output the changes in diff format and apply those.
BUF_VERSION=$(cat BUF_VERSION)
SQLC_VERSION=$(cat SQLC_VERSION)
docker run --rm -i --volume "./:/workspace" --workdir "/workspace" "bufbuild/buf:${BUF_VERSION}" format -d | patch --quiet
docker run --rm -i --volume "./:/workspace" --workdir "/workspace" "bufbuild/buf:${BUF_VERSION}" generate &
docker run --rm -i --volume "./:/workspace" --workdir "/workspace" "bufbuild/buf:${BUF_VERSION}" generate --template ./buf-es.gen.yaml &
docker run --rm -i --volume "./:/src" --workdir "/src" "sqlc/sqlc:${SQLC_VERSION}" generate &
wait

rm -rf ../frontend/neosync-api-client
mkdir -p ../frontend/neosync-api-client
mv gen/es/protos/** ../frontend/neosync-api-client
rm -rf gen/es
