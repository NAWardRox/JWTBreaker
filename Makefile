BINARY_NAME=jwt-crack
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"

.PHONY: all build clean test install demo server help

all: clean build

build:
\t@echo "Building $(BINARY_NAME)..."
\tgo build $(LDFLAGS) -o $(BINARY_NAME) .

clean:
\t@echo "Cleaning..."
\trm -f $(BINARY_NAME)
\trm -rf build/ dist/

test:
\t@echo "Running tests..."
\tgo test -v ./...

install:
\t@echo "Installing dependencies..."
\tgo mod tidy
\tgo mod download

demo: build
\t@echo "Running demo..."
\t./$(BINARY_NAME) crack --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c" --smart

server: build
\t@echo "Starting web server..."
\t./$(BINARY_NAME) serve --web --port 8080

help:
\t@echo "JWT Bruteforcer Build System"
\t@echo ""
\t@echo "Available targets:"
\t@echo "  build    - Build the binary"
\t@echo "  clean    - Clean build artifacts" 
\t@echo "  test     - Run tests"
\t@echo "  demo     - Run demo attack"
\t@echo "  server   - Start web server"
\t@echo "  install  - Install dependencies"
\t@echo "  help     - Show this help"

.DEFAULT_GOAL := help
