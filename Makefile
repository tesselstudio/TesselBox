# TesselBox Build Makefile

BINARY_NAME=tesselbox
DIST_DIR=dist
CMD_PATH=./cmd/main.go

# Build flags
LDFLAGS=-s -w

.PHONY: all clean build build-local build-linux build-windows build-macos build-android build-ios

# Detect OS and architecture for local builds
UNAME_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
UNAME_ARCH := $(shell uname -m)

# Map OS names
ifeq ($(UNAME_OS),linux)
	LOCAL_OS := linux
	LOCAL_EXT :=
else ifeq ($(UNAME_OS),darwin)
	LOCAL_OS := darwin
	LOCAL_EXT :=
else ifeq ($(findstring mingw,$(UNAME_OS)),mingw)
	LOCAL_OS := windows
	LOCAL_EXT := .exe
else ifeq ($(findstring msys,$(UNAME_OS)),msys)
	LOCAL_OS := windows
	LOCAL_EXT := .exe
else
	LOCAL_OS := linux
	LOCAL_EXT :=
endif

# Map architecture
ifeq ($(UNAME_ARCH),x86_64)
	LOCAL_ARCH := amd64
else ifeq ($(UNAME_ARCH),amd64)
	LOCAL_ARCH := amd64
else ifeq ($(UNAME_ARCH),arm64)
	LOCAL_ARCH := arm64
else ifeq ($(UNAME_ARCH),aarch64)
	LOCAL_ARCH := arm64
else
	LOCAL_ARCH := amd64
endif

LOCAL_BINARY := $(BINARY_NAME)-$(LOCAL_OS)-$(LOCAL_ARCH)$(LOCAL_EXT)

# Default: build for current platform only
# Note: Cross-compilation requires CGO setup for OpenGL/GLFW
all: clean build-local

# Create dist directory
$(DIST_DIR):
	mkdir -p $(DIST_DIR)

clean:
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)

# Build for local OS/architecture (allows CGO to work properly)
build: build-local

build-local: $(DIST_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(LOCAL_BINARY) $(CMD_PATH)
	@echo "Built for local platform: $(LOCAL_OS)/$(LOCAL_ARCH) -> $(LOCAL_BINARY)"

# Linux builds
build-linux: build-linux-amd64

build-linux-amd64: $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	@echo "Built Linux AMD64"

# Windows builds
build-windows: build-windows-amd64 build-windows-arm64

build-windows-amd64: $(DIST_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)
	@echo "Built Windows AMD64"

build-windows-arm64: $(DIST_DIR)
	GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-arm64.exe $(CMD_PATH)
	@echo "Built Windows ARM64"

# macOS builds
build-macos: build-macos-amd64 build-macos-arm64 build-macos-universal

build-macos-amd64: $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH)
	@echo "Built macOS AMD64"

build-macos-arm64: $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)
	@echo "Built macOS ARM64"

build-macos-universal: build-macos-amd64 build-macos-arm64
	@which lipo > /dev/null 2>&1 && lipo -create -output $(DIST_DIR)/$(BINARY_NAME)-darwin-universal $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 && echo "Built macOS Universal" || echo "Skipping universal binary (lipo not found)"

# Android build (requires gomobile)
build-android:
	@which gomobile > /dev/null 2>&1 || (echo "Installing gomobile..." && go install golang.org/x/mobile/cmd/gomobile@latest && gomobile init)
	gomobile build -target=android -o $(DIST_DIR)/$(BINARY_NAME).apk $(CMD_PATH)
	@echo "Built Android APK"

# iOS build (requires gomobile and macOS)
build-ios:
	@which gomobile > /dev/null 2>&1 || (echo "Installing gomobile..." && go install golang.org/x/mobile/cmd/gomobile@latest && gomobile init)
	gomobile build -target=ios -o $(DIST_DIR)/$(BINARY_NAME).app $(CMD_PATH)
	@echo "Built iOS App"

# Run locally (unified launcher with Fyne GUI + OpenGL integration)
run:
	go run ./cmd/main.go

# Run as server
run-server:
	go run $(CMD_PATH) --server

# Development build (no optimizations)
dev:
	go build -o $(DIST_DIR)/$(BINARY_NAME)-dev $(CMD_PATH)

# Test
test:
	go test -v -race -coverprofile=coverage.out ./pkg/...

# Coverage
coverage:
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Vet
vet:
	go vet ./...

# Lint
lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# All checks
check: fmt vet lint test
	@echo "All checks passed!"

# Dependencies
deps:
	go mod download
	go mod tidy

# Build all platforms
build-all: build-linux build-windows build-macos
	@echo "All builds complete!"
	@ls -lh $(DIST_DIR)/
