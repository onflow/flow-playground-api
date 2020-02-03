.PHONY: generate
generate:
	GO111MODULE=on go generate ./...

.PHONY: start
start:
	GO111MODULE=on go run server/server.go