#!/bin/bash
set -e

# TesselBox Release Build Script
# Usage: ./scripts/build-release.sh <version>
# Example: ./scripts/build-release.sh v1.0.0

VERSION=${1:-"v1.0.0"}
DIST_DIR="dist"
BINARY_NAME="tesselbox"

echo "================================"
echo "TesselBox Release Build: $VERSION"
echo "================================"

# Clean previous builds
echo "Cleaning previous builds..."
rm -rf $DIST_DIR
mkdir -p $DIST_DIR

# Build flags
LDFLAGS="-s -w -X main.Version=$VERSION"

echo ""
echo "Building for all platforms..."
echo ""

# Linux AMD64
echo "Building Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o $DIST_DIR/${BINARY_NAME}-linux-amd64 ./cmd/main.go

# Windows AMD64
echo "Building Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o $DIST_DIR/${BINARY_NAME}-windows-amd64.exe ./cmd/main.go

# macOS AMD64
echo "Building macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o $DIST_DIR/${BINARY_NAME}-darwin-amd64 ./cmd/main.go

# macOS ARM64
echo "Building macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o $DIST_DIR/${BINARY_NAME}-darwin-arm64 ./cmd/main.go

# Create Universal Binary for macOS (if lipo is available)
if command -v lipo &> /dev/null; then
    echo "Creating macOS Universal Binary..."
    lipo -create \
        $DIST_DIR/${BINARY_NAME}-darwin-amd64 \
        $DIST_DIR/${BINARY_NAME}-darwin-arm64 \
        -o $DIST_DIR/${BINARY_NAME}-darwin-universal
else
    echo "lipo not found, skipping universal binary"
fi

echo ""
echo "Creating packages..."
echo ""

# Linux package
echo "Packaging Linux..."
mkdir -p $DIST_DIR/${BINARY_NAME}-linux
cp $DIST_DIR/${BINARY_NAME}-linux-amd64 $DIST_DIR/${BINARY_NAME}-linux/${BINARY_NAME}
cp -r game_content $DIST_DIR/${BINARY_NAME}-linux/
cp README.md LICENSE $DIST_DIR/${BINARY_NAME}-linux/
cat > $DIST_DIR/${BINARY_NAME}-linux/${BINARY_NAME}.desktop << 'EOF'
[Desktop Entry]
Name=TesselBox
Comment=3D Voxel Adventure Game
Exec=tesselbox
Icon=tesselbox
Type=Application
Categories=Game;AdventureGame;
Terminal=false
EOF
tar -czf $DIST_DIR/${BINARY_NAME}-linux-amd64.tar.gz -C $DIST_DIR ${BINARY_NAME}-linux

# Windows package
echo "Packaging Windows..."
mkdir -p $DIST_DIR/${BINARY_NAME}-windows
cp $DIST_DIR/${BINARY_NAME}-windows-amd64.exe $DIST_DIR/${BINARY_NAME}-windows/${BINARY_NAME}.exe
cp -r game_content $DIST_DIR/${BINARY_NAME}-windows/
cp README.md LICENSE $DIST_DIR/${BINARY_NAME}-windows/
cd $DIST_DIR && zip -r ${BINARY_NAME}-windows-amd64.zip ${BINARY_NAME}-windows && cd ..

# macOS package (if universal binary exists)
if [ -f "$DIST_DIR/${BINARY_NAME}-darwin-universal" ]; then
    echo "Packaging macOS..."
    mkdir -p $DIST_DIR/TesselBox.app/Contents/MacOS
    mkdir -p $DIST_DIR/TesselBox.app/Contents/Resources
    cp $DIST_DIR/${BINARY_NAME}-darwin-universal $DIST_DIR/TesselBox.app/Contents/MacOS/TesselBox
    cp -r game_content $DIST_DIR/TesselBox.app/Contents/Resources/
    
    cat > $DIST_DIR/TesselBox.app/Contents/Info.plist << EOF
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
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION#v}</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
</dict>
</plist>
EOF
    
    # Create DMG if hdiutil is available (macOS only)
    if command -v hdiutil &> /dev/null; then
        hdiutil create -volname "TesselBox" -srcfolder $DIST_DIR/TesselBox.app -ov -format UDZO $DIST_DIR/TesselBox.dmg
    else
        echo "hdiutil not found, creating tar.gz instead"
        tar -czf $DIST_DIR/${BINARY_NAME}-darwin-universal.tar.gz -C $DIST_DIR TesselBox.app
    fi
fi

# Generate checksums
echo ""
echo "Generating checksums..."
cd $DIST_DIR
sha256sum *.tar.gz *.zip *.dmg 2>/dev/null > checksums.txt || sha256sum *.tar.gz *.zip > checksums.txt
cd ..

echo ""
echo "================================"
echo "Build complete!"
echo "================================"
echo ""
echo "Artifacts:"
ls -lh $DIST_DIR/
echo ""
echo "Checksums:"
cat $DIST_DIR/checksums.txt
echo ""
echo "Release $VERSION is ready!"
