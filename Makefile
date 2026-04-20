.PHONY: help build build-linux build-windows docker-build test test-unit test-integration clean

help:
	@echo "Available targets:"
	@echo "  build              - Build the tally-mcp binary"
	@echo "  build-linux        - Build for Linux"
	@echo "  build-windows      - Build for Windows"
	@echo "  docker-build       - Build Docker image"
	@echo "  test               - Run all tests (unit + integration)"
	@echo "  test-unit          - Run unit tests only"
	@echo "  test-integration   - Run integration tests (requires .env.test)"
	@echo "  clean              - Clean build artifacts"

build:
	go build -o tally-mcp .

build-linux:
	GOOS=linux GOARCH=amd64 go build -o tally-mcp .

build-windows:
	GOOS=windows GOARCH=amd64 go build -o tally-mcp.exe .

docker-build:
	docker build -f build/Dockerfile -t tally-mcp:latest .

test:
	go test -v ./...

test-unit:
	go test -v ./pkg/...

test-integration:
	@if [ ! -f .env.test ]; then \
		echo "Error: .env.test file not found."; \
		echo "Please copy .env.test.example to .env.test and update with your Tally connection details."; \
		exit 1; \
	fi
	go test -v -tags=integration ./tests/integration/...

clean:
	rm -f tally-mcp tally-mcp.exe
	go clean
