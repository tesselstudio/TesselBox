# TesselBox Build Makefile

BINARY_NAME=tesselbox
DIST_DIR=dist

# Ensure Go toolchain is on PATH
export GOROOT := /home/jason/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.10.linux-amd64
export PATH := $(GOROOT)/bin:$(PATH)

# Build flags
LDFLAGS := -s -w

.PHONY: all clean build-web run-web

all: build-web

clean:
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)

# wasm_exec.js location (Go 1.24+ uses lib/wasm; older uses misc/wasm)
WASM_EXEC := $(shell go env GOROOT)/lib/wasm/wasm_exec.js
ifeq ($(wildcard $(WASM_EXEC)),)
WASM_EXEC := $(shell go env GOROOT)/misc/wasm/wasm_exec.js
endif

# Build WebAssembly version
build-web:
	@echo "Building WebAssembly version..."
	GOOS=js GOARCH=wasm go build -ldflags="$(LDFLAGS)" -o web/main.wasm ./cmd/main_web.go
	@echo "Copying wasm_exec.js from $(WASM_EXEC)..."
	@cp -f "$(WASM_EXEC)" web/
	@echo "WebAssembly build complete! Files in web/"

# Run web version (starts server and opens browser)
run-web: build-web
	@echo "Starting local web server at http://localhost:8080"
	@(sleep 0.5; (command -v xdg-open >/dev/null && xdg-open http://localhost:8080) || \
		(command -v sensible-browser >/dev/null && sensible-browser http://localhost:8080) || \
		echo "Open http://localhost:8080 in your browser") &
	@cd web && python3 -m http.server 8080

# Dependencies
deps:
	go mod download
	go mod tidy

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

# Test
test:
	go test -v -race -coverprofile=coverage.out ./pkg/...

# Coverage
coverage:
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# All checks
check: fmt vet lint
	@echo "All checks passed!"