# Makefile for Photoptim

# Build variables
BINARY=photoptim
MAIN_DIR=cmd/photoptim

# Build the application
build:
	go build -o ${BINARY} ${MAIN_DIR}/main.go

# Install the application
install:
	go install ${MAIN_DIR}/main.go

# Run tests
test:
	go test -v ./...

# Clean build files
clean:
	rm -f ${BINARY}

# Run the application
run:
	go run ${MAIN_DIR}/main.go

# Build and install
all: build install

.PHONY: build install test clean run all