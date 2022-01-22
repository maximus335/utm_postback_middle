BUILD_DIR = dist
CMD_DIR = cmd

GO = go
LINTER = golangci-lint

MAIN_FILES = $(shell find $(CMD_DIR) -type f -name "main.go")
BINARIES = $(MAIN_FILES:$(CMD_DIR)/%/main.go=$(BUILD_DIR)/%)

default: build

$(BUILD_DIR)/%: .FORCE
	$(GO) build -o $@ $(@:$(BUILD_DIR)/%=$(CMD_DIR)/%/main.go)

download:
	$(GO) mod download

lint:
	$(LINTER) run

build: $(BINARIES)

test: download
	$(GO) test -cover ./...

coverage: download
	$(GO) test -cover -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out
	@rm coverage.out

clean:
	@rm -r $(BUILD_DIR)

.PHONY: default download lint test coverage build clean
.FORCE:
