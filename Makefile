VERSION := $(shell cat VERSION)

run:
	go run main.go

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
	git tag -a $(VERSION) -m "release $(VERSION)"
	git push origin $(VERSION)
	goreleaser release --clean

generate:
	protoc --go_out=. --go-grpc_out=. protos/cmd.proto

bench:
	go run main.go bench --num-connections=4 --engine=ironhawk
