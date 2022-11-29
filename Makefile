CURRENT_SHORT_COMMIT := $(shell git rev-parse --short HEAD)
CURRENT_VERSION := $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
LAST_KNOWN_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null)
CONTAINER := flow-playground-api
IMAGE_URL := gcr.io/dl-flow/playground-api
K8S_YAMLS_LOCATION := ./k8s
KUBECONFIG := $(shell uuidgen)
PACKAGE_TEST_DIRS = $(shell go list ./... | grep -v /e2eTest)
E2E_TEST_DIRS = $(shell go list ./e2eTest/...)

.PHONY: generate
generate:
	GO111MODULE=on go generate ./...

.PHONY: package-test-ci
package-test-ci:
	GO111MODULE=on go test -v $(PACKAGE_TEST_DIRS)

.PHONY: e2e-test-ci
e2e-test-ci:
	GO111MODULE=on go test -v $(E2E_TEST_DIRS)

.PHONY: test-local
test-local:
	GO111MODULE=on go test -v ./... -timeout 20m > test-results.log

.PHONY: run
run:
	FLOW_DEBUG=true \
	FLOW_SESSIONCOOKIESSECURE=false \
	GO111MODULE=on \
	go run \
	-ldflags "-X github.com/dapperlabs/flow-playground-api/build.version=$(LAST_KNOWN_VERSION)" \
	server/server.go

.PHONY: run-pg
run-pg:
	FLOW_DB_USER=postgres \
	FLOW_DB_PORT=5432 \
	FLOW_DB_NAME=dapper \
	FLOW_DB_HOST=localhost \
	FLOW_STORAGEBACKEND=postgresql \
	FLOW_DEBUG=true FLOW_SESSIONCOOKIESSECURE=false \
	GO111MODULE=on \
	go run \
	-ldflags "-X github.com/dapperlabs/flow-playground-api/build.version=$(LAST_KNOWN_VERSION)" \
	server/server.go

.PHONY: install-linter
install-linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin v1.47.2

.PHONY: lint
lint: check-headers
	golangci-lint run -v ./...

.PHONY: check-headers
check-headers:
	@./check-headers.sh

.PHONY: check-tidy
check-tidy: generate
	go mod tidy
	git diff --exit-code
