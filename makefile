test:
	go test -v --race ./...

test-norace:
	go test -v ./...

lint:
	golangci-lint run

mocks:
	mockery --all --keeptree

fmt-all:
	gofmt -w .