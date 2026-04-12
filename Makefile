APP_NAME=recall
BUILD_TAGS=sqlite_fts5
BINARY_NAME=recall
INSTALL_PATH=/usr/local/bin

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X recall/cmd/setup.Version=$(VERSION) \
           -X recall/cmd/setup.Commit=$(COMMIT) \
           -X recall/cmd/setup.Date=$(DATE)

build:
	go build -tags "$(BUILD_TAGS)" -ldflags "$(LDFLAGS)" -o $(APP_NAME) .

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