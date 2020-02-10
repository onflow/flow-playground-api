.PHONY: generate
generate:
	GO111MODULE=on go generate ./...

.PHONY: run
run:
	GO111MODULE=on go run server/server.go