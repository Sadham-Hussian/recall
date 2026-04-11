APP_NAME=recall
BUILD_TAGS=sqlite_fts5
BINARY_NAME=recall
INSTALL_PATH=/usr/local/bin

build:
	go build -tags "$(BUILD_TAGS)" -o $(APP_NAME) .

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)"
	sudo mv $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installed successfully."

uninstall:
	@echo "Removing $(BINARY_NAME)"
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)

test:
	go clean -testcache
	go test -tags "$(BUILD_TAGS)" ./internal/...

run:
	go run -tags "$(BUILD_TAGS)" main.go