BUF=docker run --volume "$(shell /bin/pwd):/workspace" --volume "$(PROJECT_DIR)/bin:/usr/bin" --workdir /workspace bufbuild/buf 
proto-build:
	$(BUF) build

proto-generate: deps
	$(BUF) generate

proto-lint:
	$(BUF) lint	

deps:
	$(call go-get-tool,protoc-gen-go,google.golang.org/protobuf/cmd/protoc-gen-go@latest)
	$(call go-get-tool,protoc-gen-go-grpc,google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest)

run:
	air

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

