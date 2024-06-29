.PHONY: all build clean test

all: build

install:
	@echo "Installing dependencies"
	go mod tidy

build:
	@echo "Building the Golang backend"
	go build -o aptos/services .

clean:
	@echo "Cleaning up"
	go clean
	rm -f aptos/services
test:
	@echo "Running tests"
	go test ./tests/...