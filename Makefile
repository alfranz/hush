BIN := bin/hush

.PHONY: build test integration-test lint install clean

build:
	go build -o $(BIN) ./cmd/hush

test:
	hush "go test -race ./..."

integration-test:
	go test -v -tags integration -timeout 15m ./tests/integration/...

lint:
	golangci-lint run ./...

install:
	go install ./cmd/hush

clean:
	rm -rf bin/
