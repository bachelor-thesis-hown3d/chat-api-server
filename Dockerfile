# Build the server binary
FROM golang:1.17 as builder

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

#https://skaffold.dev/docs/workflows/debug/
ARG SKAFFOLD_GO_GCFLAGS

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -gcflags="${SKAFFOLD_GO_GCFLAGS}" -a -o server cmd/server/main.go

FROM alpine
ENV GOTRACEBACK=all
WORKDIR /app
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.6 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

COPY --from=builder /workspace/server .

USER 999:999

ENTRYPOINT ["/app/server"]