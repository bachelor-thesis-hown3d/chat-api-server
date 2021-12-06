BIN="$(shell /bin/pwd)/bin"
BUF_VERSION=1.0.0-rc7
BUF=bin/buf

VERSION=0.0.2

proto-build: proto-lint
	$(BUF) build

proto-generate: deps proto-lint
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

dev:
	skaffold dev

client:
	cd cmd/client; go run client.go -host localhost

docker-build: build
  DOCKER_BUILDKIT=1 docker build --tag=$IMAGE --build-arg BUILD_ENV=builder-binary $(VERSION)

docker-push: docker-build
	docker push quay.io/hown3d/chat-api-server:$(VERSION)

build:
	go build -o _output/server ./cmd/server/main.go 

fmt:
	go fmt ./...

.PHONY: deployment
deployment:
	kubectl apply -f deployment/kube.yaml