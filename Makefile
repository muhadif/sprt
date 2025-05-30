# Makefile for sprt - sprt Client

# Variables
BINARY_NAME=sprt
MAIN_PATH=./cmd/sprt
INSTALL_PATH=/usr/local/bin
CONFIG_DIR=$(HOME)/.sprt

# Go commands
GO=go
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOTEST=$(GO) test
GOGET=$(GO) get

# Build flags
LDFLAGS=-ldflags "-s -w"

# Detect OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    OS = linux
endif
ifeq ($(UNAME_S),Darwin)
    OS = darwin
endif

.PHONY: all build clean test install uninstall

all: build

build:
	@echo "Building sprt..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_NAME)"

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...
	@echo "Tests complete"

install:
	mkdir -p $(CONFIG_DIR)
	cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete"

uninstall:
	@echo "Uninstalling sprt..."
	rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstallation complete"

# Distribution targets
dist: dist-$(OS)

dist-linux: build
	@echo "Creating Linux distribution..."
	mkdir -p dist/linux
	cp $(BINARY_NAME) dist/linux/
	cp instl.sh dist/linux/
	cp README.md dist/linux/
	tar -czf dist/sprt-linux.tar.gz -C dist/linux .
	@echo "Linux distribution created: dist/sprt-linux.tar.gz"

dist-darwin: build
	@echo "Creating macOS distribution..."
	mkdir -p dist/darwin
	cp $(BINARY_NAME) dist/darwin/
	cp instl.sh dist/darwin/
	cp README.md dist/darwin/
	tar -czf dist/sprt-darwin.tar.gz -C dist/darwin .
	@echo "macOS distribution created: dist/sprt-darwin.tar.gz"

# Help target
help:
	@echo "sprt Makefile"
	@echo "Available targets:"
	@echo "  all        - Build the application (default)"
	@echo "  build      - Build the application"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  install    - Install the application to $(INSTALL_PATH)"
	@echo "  uninstall  - Uninstall the application"
	@echo "  dist       - Create distribution package for current OS"
	@echo "  dist-linux - Create Linux distribution package"
	@echo "  dist-darwin - Create macOS distribution package"
	@echo "  help       - Show this help message"