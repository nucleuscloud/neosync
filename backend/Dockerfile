# Build the manager binary
FROM golang:1.21 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

# COPY tools.go tools.go
# COPY Makefile Makefile
# RUN make install-go-tools

# ENV PATH="$PATH:$(go env GOPATH)/bin"
# ENV GOBIN="$PWD/bin"
# ENV PATH="$PATH:$GOBIN"

# Copy in generated code
COPY gen/ gen/

# Copy the go source
COPY cmd/ cmd/
COPY internal/ internal/
COPY services/ services/

# Build
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/mgmt cmd/mgmt/*.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/bin/mgmt .
USER 65532:65532

ENTRYPOINT ["/mgmt"]

CMD ["serve", "connect"]
