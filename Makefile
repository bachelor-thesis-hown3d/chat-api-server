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

run:
	air

docker-build:
	docker build . -t quay.io/hown3d/chat-api-server:$(VERSION)

docker-push:
	docker push quay.io/hown3d/chat-api-server:$(VERSION)

fmt:
	go fmt ./...

.PHONY: deployment
deployment:
	kubectl apply -f deployment/kube.yaml