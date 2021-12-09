BIN="$(shell /bin/pwd)/bin"
BUF_VERSION=1.0.0-rc7
BUF=bin/buf

VERSION=1.0.0

proto-build: buf deps proto-lint
	$(BUF) build

proto-generate: buf deps proto-lint
	$(BUF) generate

proto-lint: buf
	$(BUF) lint	

proto-mod: buf
	cd proto && ../$(BUF) mod update

.ONESHELL:

bin_dir: 
	@mkdir bin/ &>/dev/null || true

buf: bin_dir
	@if test -f $(BUF);then echo "buf binary exists, exiting" && exit 0; fi
	curl -sSL "https://github.com/bufbuild/buf/releases/download/v$(BUF_VERSION)/buf-$(shell uname -s)-$(shell uname -m)" -o $(BUF)
	chmod +x $(BUF)

deps: bin_dir
	@mkdir bin/ &>/dev/null || true
	GOBIN=$(BIN) go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc

dev:
	skaffold dev --port-forward=pods 

debug:
	skaffold debug --port-forward=debug,pods 

client:
	cd cmd/client; go run client.go -host localhost

docker-build: 
	DOCKER_BUILDKIT=1 docker build --tag=quay.io/hown3d/chat-api-server:v$(VERSION) .

docker-push: docker-build
	docker push quay.io/hown3d/chat-api-server:v$(VERSION)

build:
	go build -o _output/server ./cmd/server/main.go 

fmt:
	go fmt ./...

.PHONY: deployment
deployment:
	kubectl apply -f deployment/kube.yaml