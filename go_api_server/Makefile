.PHONY: build test lint docker clean

BINARY_NAME=ct-explorer-go
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server/

test:
	go test -race ./...

lint:
	go vet ./...

docker:
	docker build -f ../Dockerfile.go -t ct-explorer-go ..

clean:
	rm -rf $(BUILD_DIR)
