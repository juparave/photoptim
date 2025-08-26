# Makefile for Photoptim

# Build variables
BINARY=photoptim
TUI_BINARY=photoptim-tui
MAIN_DIR=cmd/photoptim
TUI_DIR=cmd/tui

# Build the CLI application
build:
	go build -o ${BINARY} ${MAIN_DIR}/main.go

# Build the TUI application
build-tui:
	go build -o ${TUI_BINARY} ${TUI_DIR}/main.go

# Install the CLI application
install:
	go install ${MAIN_DIR}/main.go

# Run tests
test:
	go test -v ./...

# Clean build files
clean:
	rm -f ${BINARY} ${TUI_BINARY}

# Run the CLI application
run:
	go run ${MAIN_DIR}/main.go

# Run the TUI application
run-tui:
	go run ${TUI_DIR}/main.go

# Build and install
all: build install

.PHONY: build build-tui install test clean run run-tui all