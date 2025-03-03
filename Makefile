# Diretórios
SOURCE_DIR := ./types
MOCKS_DIR := internal/mocks

# Nome do binário
BINARY_NAME := myapp

# Comandos
.PHONY: all build test clean mocks

test:
	@echo "Running tests..."
	go test ./...

clean:
	@echo "Cleaning up..."
	go clean
	rm -f $(BINARY_NAME)
	rm -rf $(MOCKS_DIR)

mocks:
	@echo "Generating mocks..."
	mkdir -p $(MOCKS_DIR)
	for file in $(SOURCE_DIR)/*.go; do \
		file_name=$$(basename $$file .go); \
		mockgen -source=$$file -destination=$(MOCKS_DIR)/$${file_name}.go -package=mocks; \
	done