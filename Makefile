# Variables
BINARY_NAME=ethereum_parser
SRC_DIR=./cmd/blockchain_parser
TEST_SCRIPT=./test/integration_tests.sh
PID_FILE=./$(BINARY_NAME).pid

# Build the Go binary
build:
	@echo "Building Go binary..."
	GOOS=linux GOARCH=amd64 go build -o ./bin/$(BINARY_NAME)_linux $(SRC_DIR)
	GOOS=darwin GOARCH=arm64 go build -o ./bin/$(BINARY_NAME)_darwin $(SRC_DIR)
	@echo "Build complete."

run:
	@echo "Running Go program..."
	go run $(SRC_DIR)/main.go
	@echo "Run complete."

# Run the unit tests
test-unit:
	@echo "Running unit tests..."
	go test -v ./...
	@echo "Unit tests complete."

# Run the integration tests using the shell script
test: build
	@echo "Running binary.."
	./bin/$(BINARY_NAME)_darwin & echo $$! > $(PID_FILE)
	sleep 1
	@echo "Running tests..."
	$(TEST_SCRIPT)
	@echo "Tests complete."
	@echo "Killing binary..."
	kill -9 `cat $(PID_FILE)`
	rm -f $(PID_FILE)

# Clean up build files
clean:
	@echo "Cleaning up..."
	rm -f ./bin/$(BINARY_NAME)_darwin
	rm -f ./bin/$(BINARY_NAME)_linux
	@echo "Clean complete."

