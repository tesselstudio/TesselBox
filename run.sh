#!/bin/bash

# TesselBox Game Launcher
# This script helps configure and run the game with different UI options

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
USE_FYNE="true"
USE_VULKAN="false"
WINDOW_WIDTH="1920"
WINDOW_HEIGHT="1080"
DEBUG="false"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show help
show_help() {
    echo "TesselBox Game Launcher"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --kaiju         Use Kaiju UI (HTML-based) instead of Fyne"
    echo "  --fyne          Use Fyne UI (Modern) [default]"
    echo "  --vulkan        Enable Vulkan rendering (hardware)"
    echo "  --software      Use software rendering [default]"
    echo "  --debug         Enable debug mode"
    echo "  --width SIZE    Set window width [default: 1920]"
    echo "  --height SIZE   Set window height [default: 1080]"
    echo "  --help          Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  TESSELBOX_USE_FYNE=true|false"
    echo "  TESSELBOX_USE_VULKAN=true|false"
    echo "  TESSELBOX_WINDOW_WIDTH=1920"
    echo "  TESSELBOX_WINDOW_HEIGHT=1080"
    echo "  TESSELBOX_DEBUG=true|false"
    echo ""
    echo "Examples:"
    echo "  $0                           # Run with Fyne UI (default)"
    echo "  $0 --kaiju                   # Run with Kaiju UI"
    echo "  $0 --vulkan                  # Run with Vulkan rendering"
    echo "  $0 --debug --width 1280      # Run with debug and 1280px width"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --kaiju)
            USE_FYNE="false"
            shift
            ;;
        --fyne)
            USE_FYNE="true"
            shift
            ;;
        --vulkan)
            USE_VULKAN="true"
            shift
            ;;
        --software)
            USE_VULKAN="false"
            shift
            ;;
        --debug)
            DEBUG="true"
            shift
            ;;
        --width)
            WINDOW_WIDTH="$2"
            shift 2
            ;;
        --height)
            WINDOW_HEIGHT="$2"
            shift 2
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Set environment variables
export TESSELBOX_USE_FYNE="$USE_FYNE"
export TESSELBOX_USE_VULKAN="$USE_VULKAN"
export TESSELBOX_WINDOW_WIDTH="$WINDOW_WIDTH"
export TESSELBOX_WINDOW_HEIGHT="$WINDOW_HEIGHT"
export TESSELBOX_DEBUG="$DEBUG"

# Print configuration
print_info "=== TesselBox Configuration ==="
print_info "UI System: $([ "$USE_FYNE" = "true" ] && echo "Fyne (Modern)" || echo "Kaiju (HTML-based)")"
print_info "Rendering: $([ "$USE_VULKAN" = "true" ] && echo "Vulkan (Hardware)" || echo "Software (CPU)")"
print_info "Window Size: ${WINDOW_WIDTH}x${WINDOW_HEIGHT}"
print_info "Debug Mode: $DEBUG"
print_info "================================"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go first."
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "Please run this script from the TesselBox root directory."
    exit 1
fi

# Build the game
print_info "Building TesselBox..."
if ! go build .; then
    print_error "Build failed!"
    exit 1
fi

print_success "Build completed!"

# Set up Vulkan if needed
if [ "$USE_VULKAN" = "false" ]; then
    print_info "Setting up software rendering..."
    export VK_ICD_FILENAMES="/usr/share/vulkan/icd.d/llvmpipe_icd.x86_64.json"
fi

# Run the game
print_info "Starting TesselBox with Fyne UI..."
print_info "Press Ctrl+C to exit"
echo ""

# Run the game with Fyne UI (actual game, not test mode)
if [ "$USE_FYNE" = "true" ]; then
    print_info "🎮 Launching REAL game with Fyne UI (not test mode)"
    go run ./cmd/fyne_game
else
    # Run the original game (test mode)
    if [ -f "./TesselBox" ]; then
        ./TesselBox
    else
        go run .
    fi
fi
