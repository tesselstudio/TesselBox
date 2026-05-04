# Build & Deployment Guide

Instructions for building, packaging, and deploying TesselBox.

## Prerequisites

### Required

- **Go 1.25+** - Download from [golang.org](https://golang.org/dl/)
- **Git** - For cloning submodules
- **Vulkan SDK** - Graphics drivers with Vulkan support

### Platform-Specific

**Linux:**
```bash
# Ubuntu/Debian
sudo apt-get install libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install mesa-libGL-devel libX11-devel
```

**Windows:**
- Visual Studio 2019+ (C++ build tools)
- Windows SDK

**macOS:**
- Xcode Command Line Tools
- macOS 10.15+ (Catalina or newer)

## Quick Start

### Clone Repository

```bash
git clone --recursive https://github.com/tesselstudio/TesselBox.git
cd TesselBox
```

If you already cloned without `--recursive`:

```bash
git submodule update --init --recursive
```

### Build

```bash
# Development build
go build -o tesselbox ./cmd/main.go

# Or use Make
make dev
```

### Run

```bash
./tesselbox
```

## Build Configurations

### Development Build

Fast compilation, debug symbols:

```bash
go build -o tesselbox-dev ./cmd/main.go
```

### Release Build

Optimized, stripped binaries:

```bash
go build -ldflags="-s -w" -o tesselbox ./cmd/main.go
```

### Debug Build

With debug information for GDB:

```bash
go build -gcflags="all=-N -l" -o tesselbox-debug ./cmd/main.go
```

## Cross-Compilation

### Using Make

```bash
# Build all platforms
make build-all

# Individual platforms
make build-linux-amd64
make build-windows-amd64
make build-macos-amd64
make build-macos-arm64
```

### Manual Cross-Compilation

**Linux AMD64:**
```bash
GOOS=linux GOARCH=amd64 go build -o dist/tesselbox-linux-amd64 ./cmd/main.go
```

**Windows AMD64:**
```bash
GOOS=windows GOARCH=amd64 go build -o dist/tesselbox-windows-amd64.exe ./cmd/main.go
```

**macOS AMD64:**
```bash
GOOS=darwin GOARCH=amd64 go build -o dist/tesselbox-darwin-amd64 ./cmd/main.go
```

**macOS ARM64 (Apple Silicon):**
```bash
GOOS=darwin GOARCH=arm64 go build -o dist/tesselbox-darwin-arm64 ./cmd/main.go
```

**Universal Binary (macOS):**
```bash
# Build both architectures, then combine
lipo -create \
  dist/tesselbox-darwin-amd64 \
  dist/tesselbox-darwin-arm64 \
  -output dist/tesselbox-darwin-universal
```

## Mobile Builds

### Android

Requires Android SDK and gomobile:

```bash
# Install gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

# Build APK
make build-android
# OR
gomobile build -target=android -o dist/tesselbox.apk ./cmd/main.go
```

### iOS

Requires macOS and Xcode:

```bash
# Build iOS app
make build-ios
# OR
gomobile build -target=ios -o dist/TesselBox.app ./cmd/main.go
```

## Packaging

### Linux (tar.gz)

```bash
make build-linux-amd64
mkdir -p dist/tesselbox-linux
cp dist/tesselbox-linux-amd64 dist/tesselbox-linux/tesselbox
cp -r game_content dist/tesselbox-linux/
cp README.md LICENSE dist/tesselbox-linux/
tar -czf dist/tesselbox-linux-amd64.tar.gz -C dist tesselbox-linux
```

### Windows (ZIP)

```bash
make build-windows-amd64
mkdir -p dist/tesselbox-windows
cp dist/tesselbox-windows-amd64.exe dist/tesselbox-windows/tesselbox.exe
cp -r game_content dist/tesselbox-windows/
cp README.md LICENSE dist/tesselbox-windows/
cd dist && zip -r tesselbox-windows-amd64.zip tesselbox-windows
```

### macOS (DMG)

```bash
make build-macos-universal
mkdir -p dist/TesselBox.app/Contents/MacOS
cp dist/tesselbox-darwin-universal dist/TesselBox.app/Contents/MacOS/TesselBox
cp -r game_content dist/TesselBox.app/Contents/Resources/

# Create Info.plist
cat > dist/TesselBox.app/Contents/Info.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>TesselBox</string>
    <key>CFBundleIdentifier</key>
    <string>com.tesselstudio.tesselbox</string>
    <key>CFBundleName</key>
    <string>TesselBox</string>
    <key>CFBundleVersion</key>
    <string>1.0.0</string>
</dict>
</plist>
EOF

# Create DMG
hdiutil create -volname "TesselBox" -srcfolder dist/TesselBox.app -ov -format UDZO dist/TesselBox.dmg
```

## CI/CD

### GitHub Actions

The repository includes workflows for:

- `.github/workflows/build.yml` - Build on push/PR
- `.github/workflows/release.yml` - Create releases on tags
- `.github/workflows/test.yml` - Run test suite

### Creating a Release

1. Update version in code
2. Commit and push
3. Create and push a tag:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

4. GitHub Actions automatically:
   - Builds for all platforms
   - Creates release with binaries
   - Generates changelog

### Local Release

```bash
# Full release build locally
./scripts/build-release.sh v1.0.0
```

## Distribution Platforms

### itch.io

```bash
# Upload to itch.io (using butler)
butler push dist/tesselbox-linux-amd64.tar.gz tesselstudio/tesselbox:linux-amd64 --userversion 1.0.0
butler push dist/tesselbox-windows-amd64.zip tesselstudio/tesselbox:windows-amd64 --userversion 1.0.0
butler push dist/TesselBox.dmg tesselstudio/tesselbox:macos --userversion 1.0.0
```

### Steam

Use Steamworks SDK and `steamcmd`:

```bash
# Build depot manifests
# See scripts/steam/ directory for examples
```

## Troubleshooting

### Vulkan Not Found

**Linux:**
```bash
# Install Vulkan drivers
sudo apt-get install mesa-vulkan-drivers vulkan-tools

# Or for NVIDIA:
sudo apt-get install nvidia-driver-xxx

# Verify:
vulkaninfo
```

**Windows:**
- Update GPU drivers from manufacturer
- Install Vulkan Runtime from LunarG

**macOS:**
```bash
# Install MoltenVK
brew install molten-vk
```

### Build Errors

**"kaijuengine.com not found":**
```bash
# Ensure submodule is initialized
git submodule update --init --recursive
```

**"undefined: hid.KeyboardKeyW":**
- Update Kaiju Engine submodule: `cd kaiju && git pull origin main`

**Linker errors on Windows:**
- Ensure Visual Studio C++ tools are installed
- Run from "Developer Command Prompt"

### Runtime Issues

**Black screen:**
- Check Vulkan support: `vulkaninfo`
- Try software renderer: `VK_ICD_FILENAMES=/usr/share/vulkan/icd.d/llvmpipe_icd.x86_64.json ./tesselbox`

**Crash on startup:**
- Check game_content directory exists
- Verify asset files are present
- Run with `--debug` flag for logs

## Development Tips

### Hot Reload

For faster development iteration:

```bash
# Use air for auto-rebuild (if supported by your setup)
air

# Or simple loop
while true; do
    go build -o tesselbox ./cmd/main.go && ./tesselbox
    sleep 1
done
```

### Profiling

```bash
# CPU profile
go build -o tesselbox ./cmd/main.go
./tesselbox -cpuprofile=cpu.prof

# Analyze
go tool pprof cpu.prof
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./pkg/world/...
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `TESSBOX_DEBUG` | Enable debug logging |
| `TESSBOX_FULLSCREEN` | Start in fullscreen |
| `TESSBOX_WIDTH` | Window width |
| `TESSBOX_HEIGHT` | Window height |
| `VK_ICD_FILENAMES` | Vulkan ICD config path |

## Docker Build

For reproducible builds:

```dockerfile
FROM golang:1.25

RUN apt-get update && apt-get install -y libgl1-mesa-dev xorg-dev

WORKDIR /app
COPY . .
RUN git submodule update --init --recursive
RUN go build -o tesselbox ./cmd/main.go

CMD ["./tesselbox"]
```

```bash
docker build -t tesselbox .
docker run -e DISPLAY=$DISPLAY -v /tmp/.X11-unix:/tmp/.X11-unix tesselbox
```
