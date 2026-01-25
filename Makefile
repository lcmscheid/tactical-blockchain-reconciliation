BINARY_NAME=simulator
CMD_PATH=cmd/simulator/main.go

# Default command
all: clean fmt test build

# Build the executable
build:
	@echo "Building the simulator..."
	@mkdir -p bin
	go build -o bin/$(BINARY_NAME) $(CMD_PATH)

# Run the simulator
run: build
	@echo "Running simulation..."
	@./bin/$(BINARY_NAME)

# Run unit tests
test:
	@echo "Running unit tests..."
	go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f bin/$(BINARY_NAME)
	@go clean

.PHONY: all build run test fmt clean
