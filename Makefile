BINARY     := lingua-cli
VERSION    := $(shell grep 'version = ' main.go | head -1 | sed 's/.*"\(.*\)".*/\1/')
MODULE     := $(shell go env GOMODCACHE 2>/dev/null; head -1 go.mod | awk '{print $$2}')
BUILD_DIR  := dist
LDFLAGS    := -s -w -X main.version=$(VERSION)

# Target platforms: OS/ARCH pairs
PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	linux/arm \
	windows/amd64 \
	windows/arm64

.PHONY: all build clean release tidy fmt vet test help

## all: Build for the current platform
all: build

## build: Build for the current platform
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

## tidy: Tidy go modules
tidy:
	go mod tidy

## fmt: Format source code
fmt:
	gofmt -w .

## vet: Run go vet
vet:
	go vet ./...

## test: Run tests
test:
	go test ./...

## release: Cross-compile for all target platforms and create archives
release: tidy
	@mkdir -p $(BUILD_DIR)
	@$(foreach PLATFORM,$(PLATFORMS),$(MAKE) _build_platform PLATFORM=$(PLATFORM);)
	@echo ""
	@echo "Release archives created in $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/

_build_platform:
	$(eval GOOS   := $(word 1,$(subst /, ,$(PLATFORM))))
	$(eval GOARCH := $(word 2,$(subst /, ,$(PLATFORM))))
	$(eval SUFFIX := $(if $(filter windows,$(GOOS)),.exe,))
	$(eval OUTBIN := $(BUILD_DIR)/$(BINARY)-$(VERSION)-$(GOOS)-$(GOARCH)$(SUFFIX))
	$(eval ARCNAME:= $(BINARY)-$(VERSION)-$(GOOS)-$(GOARCH))
	@echo "Building $(GOOS)/$(GOARCH)..."
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o $(OUTBIN) .
	@if [ "$(GOOS)" = "windows" ]; then \
		cd $(BUILD_DIR) && zip -q $(ARCNAME).zip $(BINARY)-$(VERSION)-$(GOOS)-$(GOARCH)$(SUFFIX) && \
		rm -f $(BINARY)-$(VERSION)-$(GOOS)-$(GOARCH)$(SUFFIX); \
	else \
		cd $(BUILD_DIR) && tar czf $(ARCNAME).tar.gz $(BINARY)-$(VERSION)-$(GOOS)-$(GOARCH)$(SUFFIX) && \
		rm -f $(BINARY)-$(VERSION)-$(GOOS)-$(GOARCH)$(SUFFIX); \
	fi
	@echo "  -> $(BUILD_DIR)/$(ARCNAME)$(if $(filter windows,$(GOOS)),.zip,.tar.gz)"

## checksums: Generate SHA256 checksums for release archives
checksums:
	@cd $(BUILD_DIR) && sha256sum *.tar.gz *.zip 2>/dev/null > checksums.sha256
	@echo "Checksums written to $(BUILD_DIR)/checksums.sha256"

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR) $(BINARY)

## install: Install to GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" .

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^##' Makefile | sed 's/## /  /'
