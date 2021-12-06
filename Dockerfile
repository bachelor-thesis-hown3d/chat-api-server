ARG BUILD_ENV=builder-golang

# Build the server binary
FROM golang:1.16 as builder-golang

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/server cmd/server
COPY pkg/ pkg/
COPY proto/ proto/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o server cmd/server/main.go

FROM scratch as builder-binary
WORKDIR /workspace
COPY _output/server /workspace/server

FROM ${BUILD_ENV} as build

FROM alpine
WORKDIR /app
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.6 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

COPY --from=build /workspace/server .

USER 999:999

ENTRYPOINT ["/app/server"]