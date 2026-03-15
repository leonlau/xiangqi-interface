GO ?= go
ENGINE_DIR ?= $(CURDIR)/vendor/engine
PLUGIN_DIR ?= $(CURDIR)/plugin
DIST_DIR ?= $(CURDIR)/dist
SO_NAME ?= libengine.so

.PHONY: all build-plugin test test-integration ci clean

all: build-plugin

build-plugin:
	@if [ ! -e "$(ENGINE_DIR)/go.mod" ]; then \
		echo "Error: engine source not found at $(ENGINE_DIR)."; \
		echo "Set ENGINE_DIR env var or populate vendor/engine/ before running make."; \
		exit 1; \
	fi
	mkdir -p $(DIST_DIR)
	cd $(PLUGIN_DIR) && CGO_ENABLED=1 $(GO) build -buildmode=plugin -o $(DIST_DIR)/$(SO_NAME) .

test:
	cd $(CURDIR) && $(GO) vet ./...
	cd $(PLUGIN_DIR) && $(GO) test ./...
	cd $(CURDIR)/sdk && $(GO) test ./...

test-integration: build-plugin
	cd $(CURDIR)/sdk && CGO_ENABLED=1 ENGINE_PLUGIN=$(CURDIR)/dist/$(SO_NAME) $(GO) test -tags=integration ./...

ci: test build-plugin test-integration

clean:
	rm -rf $(DIST_DIR)