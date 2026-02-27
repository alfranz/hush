BIN := bin/hush

.PHONY: build test lint install clean

build:
	go build -o $(BIN) ./cmd/hush

test:
	go test -race ./...

lint:
	golangci-lint run ./...

install:
	go install ./cmd/hush

clean:
	rm -rf bin/
