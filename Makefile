.PHONY: build test clean run install

# Build the binary
build:
	go build -v -o anki-mcp .

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f anki-mcp

# Run the server
run: build
	./anki-mcp

# Install dependencies
deps:
	go mod tidy
	go mod download

# Install the binary
install:
	go install .

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o anki-mcp-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o anki-mcp-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o anki-mcp-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o anki-mcp-windows-amd64.exe .

# Check if AnkiConnect is available
check-anki:
	@echo "Checking AnkiConnect availability..."
	@curl -s -X POST http://localhost:8765 -H "Content-Type: application/json" -d '{"action": "version", "version": 6}' | grep -q '"result"' && echo "✓ AnkiConnect is available" || echo "✗ AnkiConnect is not available"

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo "  run        - Build and run the server"
	@echo "  deps       - Install dependencies"
	@echo "  install    - Install the binary"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  check-anki - Check AnkiConnect availability"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code"
	@echo "  help       - Show this help"