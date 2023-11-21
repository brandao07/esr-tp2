BINARY_NAME = main

BUILD_FLAGS = -o ./cmd/$(BINARY_NAME)

# Build the executable
build:
	
	@go fmt ./...
	@echo "Building $(BINARY_NAME)..."
	@go build $(BUILD_FLAGS) ./src

# Run the application with arguments
run:
	@mkdir -p ./out
	@echo "Running $(BINARY_NAME) with arguments: $(ARGS)...\n"
	@./cmd/$(BINARY_NAME) $(ARGS)

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f ./cmd/$(BINARY_NAME)
	@rm -d ./cmd
	@rm -rf ./out

.PHONY: build run clean all