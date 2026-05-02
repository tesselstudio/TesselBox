#!/bin/bash

# TesselBox Cross-Platform Build Script

set -e

BINARY_NAME="tesselbox"
DIST_DIR="dist"
CMD_PATH="./cmd/main.go"
LDFLAGS="-s -w"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create dist directory
mkdir -p "$DIST_DIR"

# Function to print status
print_status() {
    echo -e "${GREEN}[BUILD]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Build Linux
build_linux() {
    print_status "Building for Linux..."
    
    # AMD64
    GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$DIST_DIR/${BINARY_NAME}-linux-amd64" "$CMD_PATH"
    print_status "Linux AMD64 build complete"
    
    # ARM64
    GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "$DIST_DIR/${BINARY_NAME}-linux-arm64" "$CMD_PATH"
    print_status "Linux ARM64 build complete"
}

# Build Windows
build_windows() {
    print_status "Building for Windows..."
    
    # AMD64
    GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$DIST_DIR/${BINARY_NAME}-windows-amd64.exe" "$CMD_PATH"
    print_status "Windows AMD64 build complete"
    
    # ARM64
    GOOS=windows GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "$DIST_DIR/${BINARY_NAME}-windows-arm64.exe" "$CMD_PATH"
    print_status "Windows ARM64 build complete"
}

# Build macOS
build_macos() {
    print_status "Building for macOS..."
    
    # AMD64
    GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$DIST_DIR/${BINARY_NAME}-darwin-amd64" "$CMD_PATH"
    print_status "macOS AMD64 build complete"
    
    # ARM64
    GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "$DIST_DIR/${BINARY_NAME}-darwin-arm64" "$CMD_PATH"
    print_status "macOS ARM64 build complete"
    
    # Universal binary (if lipo is available)
    if command -v lipo &> /dev/null; then
        lipo -create -output "$DIST_DIR/${BINARY_NAME}-darwin-universal" \
            "$DIST_DIR/${BINARY_NAME}-darwin-amd64" \
            "$DIST_DIR/${BINARY_NAME}-darwin-arm64" || true
        print_status "macOS Universal binary created"
    else
        print_warning "lipo not found, skipping universal binary"
    fi
}

# Build Android
build_android() {
    print_status "Building for Android..."
    
    if ! command -v gomobile &> /dev/null; then
        print_status "Installing gomobile..."
        go install golang.org/x/mobile/cmd/gomobile@latest
        gomobile init
    fi
    
    gomobile build -target=android -o "$DIST_DIR/${BINARY_NAME}.apk" "$CMD_PATH"
    print_status "Android APK build complete"
}

# Build iOS
build_ios() {
    print_status "Building for iOS..."
    
    if ! command -v gomobile &> /dev/null; then
        print_status "Installing gomobile..."
        go install golang.org/x/mobile/cmd/gomobile@latest
        gomobile init
    fi
    
    if [[ "$OSTYPE" != "darwin"* ]]; then
        print_error "iOS builds require macOS"
        return 1
    fi
    
    gomobile build -target=ios -o "$DIST_DIR/${BINARY_NAME}.app" "$CMD_PATH"
    print_status "iOS App build complete"
}

# Clean build artifacts
clean() {
    print_status "Cleaning build artifacts..."
    rm -rf "$DIST_DIR"
    mkdir -p "$DIST_DIR"
}

# Show help
show_help() {
    echo "TesselBox Build Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  all       Build for all platforms (default)"
    echo "  linux     Build for Linux (amd64, arm64)"
    echo "  windows   Build for Windows (amd64, arm64)"
    echo "  macos     Build for macOS (amd64, arm64, universal)"
    echo "  android   Build Android APK"
    echo "  ios       Build iOS App (macOS only)"
    echo "  clean     Clean build artifacts"
    echo "  help      Show this help message"
    echo ""
}

# Main script
main() {
    case "${1:-all}" in
        all)
            clean
            build_linux
            build_windows
            build_macos
            print_status "All builds complete!"
            ls -lh "$DIST_DIR/"
            ;;
        linux)
            build_linux
            ;;
        windows)
            build_windows
            ;;
        macos)
            build_macos
            ;;
        android)
            build_android
            ;;
        ios)
            build_ios
            ;;
        clean)
            clean
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
