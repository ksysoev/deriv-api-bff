test:
	go test -v --race ./...

test-norace:
	go test -v ./...

lint:
	golangci-lint run

mocks:
	mockery --all

fmt-all:
	gofmt -w .

build:
	go build ./...
	go install ./...

mod-tidy:
	go mod tidy

coverage:
	go test ./... -coverprofile=cover.out
