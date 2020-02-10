.PHONY: generate
generate:
	GO111MODULE=on go generate ./...

.PHONY: test
test:
	GO111MODULE=on go test ./...

.PHONY: run
run:
	GO111MODULE=on go run server/server.go
