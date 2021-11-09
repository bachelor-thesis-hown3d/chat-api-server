BIN="$(shell /bin/pwd)/bin"
BUF_VERSION=1.0.0-rc7
BUF=bin/buf
proto-build:
	$(BUF) build

proto-generate: deps
	$(BUF) generate

proto-lint:
	$(BUF) lint	

buf:
	@mkdir bin/ || true
	curl -sSL "https://github.com/bufbuild/buf/releases/download/v$(BUF_VERSION)/buf-$(shell uname -s)-$(shell uname -m)" -o $(BUF)
	chmod +x $(BUF)

deps:
	@mkdir bin/ || true
	GOBIN=$(BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	GOBIN=$(BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

run:
	air

fmt:
	go fmt ./...
