# CARES Project Makefile
# Handles protocol buffer generation and project building

.PHONY: proto clean build run-orchestrator run-worker help

# Default target
all: proto build

# Generate Go code from protocol buffer definitions
proto:
	@echo "Generating gRPC code from protobuf..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/cluster/cluster.proto
	@echo "✅ Protocol buffer code generated successfully"

# Build the binary
build:
	@echo "Building CARES binary..."
	@go build -o bin/cares cmd/cares/main.go
	@echo "✅ Binary built successfully at bin/cares"

# Clean generated files and build artifacts
clean:
	@echo "Cleaning generated files..."
	@rm -f internal/cluster/*.pb.go
	@rm -f bin/cares
	@echo "✅ Cleaned successfully"

# Run in orchestrator mode (for development)
run-orchestrator: build
	@echo "Starting CARES in orchestrator mode..."
	@./bin/cares

# Run in worker mode (for development)
run-worker: build
	@echo "Starting CARES in worker mode..."
	@./bin/cares

# Install required tools
install-tools:
	@echo "Installing required tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "✅ Tools installed successfully!"
	@echo "If 'make proto' still fails, run: export PATH=\$$PATH:\$$HOME/go/bin"

# Update dependencies
deps:
	@echo "Updating dependencies..."
	@go mod tidy
	@go mod download
	@echo "✅ Dependencies updated"

# Help target
help:
	@echo "CARES Project Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  proto              Generate gRPC code from protobuf definitions"
	@echo "  build              Build the CARES binary"
	@echo "  clean              Clean generated files and build artifacts"
	@echo "  run-orchestrator   Build and run in orchestrator mode"
	@echo "  run-worker         Build and run in worker mode"
	@echo "  install-tools      Install required protoc plugins"
	@echo "  deps               Update Go dependencies"
	@echo "  help               Show this help message"
	@echo ""
	@echo "Example usage:"
	@echo "  make proto build   # Generate protobuf code and build binary"
	@echo "  make clean all     # Clean and rebuild everything"
