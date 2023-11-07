BINARY_NAME = main

BUILD_FLAGS = -o ./cmd/$(BINARY_NAME)

# Build the executable
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(BUILD_FLAGS) ./src

# Run the application with arguments
run:
	@echo "Running $(BINARY_NAME) with arguments: $(ARGS)...\n"
	@./cmd/$(BINARY_NAME) $(ARGS)

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f ./cmd/$(BINARY_NAME)
	@rm -d ./cmd

# Build and run the application
all: build run

.PHONY: build run clean all