SHORT_COMMIT := $(shell git rev-parse --short HEAD)

.PHONY: generate
generate:
	GO111MODULE=on go generate ./...

.PHONY: test
test:
	GO111MODULE=on go test ./...

.PHONY: run
run:
	GO111MODULE=on go run server/server.go

.PHONY: docker-build
docker-build:
	DOCKER_BUILDKIT=1 docker build --ssh default -t gcr.io/dl-flow/playground-api:latest -t "gcr.io/dl-flow/playground-api:$(SHORT_COMMIT)" .

.PHONY: docker-push
docker-push:
	docker push gcr.io/dl-flow/playground-api:latest
	docker push "gcr.io/dl-flow/playground-api:$(SHORT_COMMIT)"
