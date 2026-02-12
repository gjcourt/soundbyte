# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_SERVER=server
BINARY_CLIENT=client
BINARY_UNIX_SERVER=$(BINARY_SERVER)_unix
BINARY_UNIX_CLIENT=$(BINARY_CLIENT)_unix

# Linting
LINTER=golangci-lint

all: test build

build:
	$(GOBUILD) -o $(BINARY_SERVER) -v ./cmd/server
	$(GOBUILD) -o $(BINARY_CLIENT) -v ./cmd/client

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_SERVER) $(BINARY_CLIENT)
	rm -f $(BINARY_UNIX_SERVER) $(BINARY_UNIX_CLIENT)

run-server:
	$(GOBUILD) -o $(BINARY_SERVER) -v ./cmd/server
	./$(BINARY_SERVER)

run-client:
	$(GOBUILD) -o $(BINARY_CLIENT) -v ./cmd/client
	./$(BINARY_CLIENT)

deps:
	$(GOGET) -v ./...

# Linting target
# Requires golangci-lint installed: https://golangci-lint.run/usage/install/
lint:
	$(LINTER) run ./...

# Docker
docker-build:
	docker-compose build

docker-up:
	docker-compose up

.PHONY: all build test clean run-server run-client deps lint docker-build docker-up
