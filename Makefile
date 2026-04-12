.PHONY: help build build-linux build-windows docker-build test clean

help:
	@echo "Available targets: build, build-linux, build-windows, docker-build, test, clean"

build:
	go build -o tally-mcp .

build-linux:
	GOOS=linux GOARCH=amd64 go build -o tally-mcp .

build-windows:
	GOOS=windows GOARCH=amd64 go build -o tally-mcp.exe .

docker-build:
	docker build -t tally-mcp:latest .

test:
	go test -v ./...

clean:
	rm -f tally-mcp tally-mcp.exe
	go clean
