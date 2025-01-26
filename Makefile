build:
	go build -o ./dicedb-cli

check-golangci-lint:
	@if ! command -v golangci-lint > /dev/null || ! golangci-lint version | grep -q "$(GOLANGCI_LINT_VERSION)"; then \
		echo "Required golangci-lint version $(GOLANGCI_LINT_VERSION) not found."; \
		echo "Please install golangci-lint version $(GOLANGCI_LINT_VERSION) with the following command:"; \
		echo "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.60.1"; \
		exit 1; \
	fi

lint: check-golangci-lint
	golangci-lint run ./...

release:
	git tag v0.0.3
	git push origin v0.0.3
	goreleaser release --clean

generate:
	protoc --go_out=. --go-grpc_out=. protos/cmd.proto

update:
	git submodule update --remote
	git submodule update --init --recursive
