CURRENT_SHORT_COMMIT := $(shell git rev-parse --short HEAD)
CURRENT_VERSION := $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
LAST_KNOWN_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null)
CONTAINER := flow-playground-api
IMAGE_URL := gcr.io/dl-flow/playground-api
K8S_YAMLS_LOCATION := ./k8s
KUBECONFIG := $(shell uuidgen)

.PHONY: generate
generate:
	GO111MODULE=on go generate ./...

.PHONY: test
test:
	GO111MODULE=on go test -v ./...

.PHONY: test-datastore
test-datastore:
	DATASTORE_EMULATOR_HOST=localhost:8081 FLOW_STORAGEBACKEND=datastore GO111MODULE=on go test ./...

.PHONY: run
run:
	FLOW_DEBUG=true \
	FLOW_SESSIONCOOKIESSECURE=false \
	GO111MODULE=on \
	go run \
	-ldflags "-X github.com/dapperlabs/flow-playground-api/build.version=$(LAST_KNOWN_VERSION)" \
	server/server.go

.PHONY: run-datastore
run-datastore:
	DATASTORE_EMULATOR_HOST=localhost:8081 \
	FLOW_STORAGEBACKEND=datastore \
	FLOW_DATASTORE_GCPPROJECTID=flow-developer-playground \
	FLOW_DEBUG=true FLOW_SESSIONCOOKIESSECURE=false \
	GO111MODULE=on \
	go run \
	-ldflags "-X github.com/dapperlabs/flow-playground-api/build.version=$(LAST_KNOWN_VERSION)" \
	server/server.go

.PHONY: start-datastore-emulator
start-datastore-emulator:
	gcloud beta emulators datastore start --no-store-on-disk

.PHONY: ci
ci: check-tidy test check-headers

.PHONY: install-linter
install-linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin v1.47.2

.PHONY: lint
lint:
	golangci-lint run -v ./...

.PHONY: check-headers
check-headers:
	@./check-headers.sh

.PHONY: check-tidy
check-tidy: generate
	go mod tidy
	git diff --exit-code
