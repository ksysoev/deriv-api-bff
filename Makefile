test:
	go test -v --race ./...

test-norace:
	go test -v ./...

lint:
	golangci-lint run

mocks:
	mockery

fmt-all:
	gofmt -w .

build:
	go build ./...

install:
	go install ./...
	cp ./scripts/pre-commit ./.git/hooks/pre-commit

mod-tidy:
	go mod tidy

coverage:
	go test ./... -coverprofile=cover.out

profile:
	go test -benchmem -run=^$$ -benchtime=5s -bench ^Benchmark -cpuprofile=cpu.out -memprofile=mem.out  -blockprofile=lock.out ./tests

bench:
	go test -benchmem -run=^$$ -benchtime=5s -bench ^Benchmark ./tests
