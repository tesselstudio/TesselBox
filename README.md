# TesselBox

A hexagonal block-based adventure game with exploration, crafting, combat, survival elements, and **multiplayer support**.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build](https://github.com/tesselstudio/TesselBox/actions/workflows/build.yml/badge.svg)](https://github.com/tesselstudio/TesselBox/actions)

## Overview

TesselBox is an open-source game built with a unique hexagonal grid system. Explore vast worlds, gather resources, craft tools, build structures, survive against hostile mobs, and **play with friends** in multiplayer mode.

## Features

- **Hexagonal Grid System** - Unique 6-sided block world
- **Procedural Generation** - Infinite worlds with varied biomes
- **Crafting System** - Craft tools, weapons, and building materials
- **Survival Mode** - Manage health, hunger, and fend off enemies
- **Creative Mode** - Build without limits
- **Dimensions** - Travel between overworld and Randomland
- **Multiplayer** - Play with up to 16 players via LAN or direct connection
- **Cross-Platform** - Play on PC (Windows/Linux/macOS) or Mobile (Android/iOS)

## Multiplayer

TesselBox now supports multiplayer gaming:

- **UDP** for real-time position updates (50ms sync)
- **TCP** for reliable messaging (chat, block changes)
- **LAN Discovery** - Find servers automatically on your local network
- **Direct Connect** - Connect by IP address
- **Up to 16 players** per server
- **PC can host** - Run as dedicated server or host while playing
- **Mobile support** - Mobile clients can connect to PC servers

### Server Commands (PC)

```bash
# Run as dedicated server
./tesselbox --server --port 25565 --name "My Server"

# Connect to a server
./tesselbox --connect 192.168.1.100:25565 --player "Player1"

# Enable LAN discovery
./tesselbox --discover
```

## Building

### Requirements

- Go 1.21 or later
- For Android: Android SDK and NDK
- For iOS: macOS with Xcode

### Quick Build

```bash
# Clone the repository
git clone https://github.com/tesselstudio/TesselBox.git
cd TesselBox

# Build for current platform
go build -o tesselbox ./cmd/main.go

# Or use make to build all platforms
make build-all
```

### Platform-Specific Builds

```bash
# Linux
make build-linux

# Windows
make build-windows

# macOS
make build-macos

# Android APK
make build-android

# iOS
make build-ios
```

### Build Script

```bash
# Build all platforms
./scripts/build.sh all

# Build specific platform
./scripts/build.sh linux
./scripts/build.sh android
```

## Running

```bash
# Run the game
./tesselbox

# Run as dedicated server
./tesselbox --server

# Connect to a server
./tesselbox --connect 192.168.1.100:25565
```

## Controls

### PC/Desktop
| Key | Action |
|-----|--------|
| WASD | Move |
| Space | Jump |
| Shift | Sprint |
| Mouse | Look/Mine/Place |
| I | Inventory |
| C | Crafting |
| B | Block Library |

### Mobile
| Gesture | Action |
|---------|--------|
| Swipe (left half) | Move/Jump |
| Tap (right half) | Mine/Attack |
| Hold (right half) | Place block |
| Pinch | Zoom |
| Two-finger tap | Inventory |

## Project Structure

```
TesselBox/
├── cmd/                    # Entry points
│   └── main.go            # Main entry (desktop + mobile)
├── internal/
│   └── game/              # Core game logic
├── pkg/
│   ├── network/           # Multiplayer networking
│   │   ├── packet.go      # Packet serialization
│   │   ├── udp.go         # UDP transport
│   │   ├── tcp.go         # TCP transport
│   │   ├── server.go      # Game server
│   │   ├── client.go      # Network client
│   │   ├── discovery.go   # LAN discovery
│   │   └── client_session.go
│   ├── platform/          # Platform-specific code
│   │   ├── desktop.go     # PC platform
│   │   └── mobile.go      # Mobile platform
│   ├── world/             # World generation
│   ├── player/            # Player logic
│   ├── blocks/            # Block definitions
│   ├── crafting/          # Crafting system
│   └── ...
├── config/                # Configuration files
├── scripts/               # Build scripts
├── .github/workflows/      # CI/CD
├── Makefile              # Build automation
└── go.mod                # Go module
```

## Documentation

- [Contributing Guidelines](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Security Policy](SECURITY.md)

## Community

- [Discussions](https://github.com/tesselstudio/TesselBox/discussions) - Ask questions, share ideas
- [Issues](https://github.com/tesselstudio/TesselBox/issues) - Report bugs, request features

## Development Status

| Platform | Status | Multiplayer |
|----------|--------|-------------|
| PC (Windows/Linux/macOS) | Active | Host + Client |
| Mobile (Android/iOS) | Active | Client only |

## CI/CD

The project uses GitHub Actions for automated builds:

- **Linux**: AMD64, ARM64
- **Windows**: AMD64, ARM64
- **macOS**: AMD64, ARM64, Universal
- **Android**: APK
- **iOS**: App bundle

Releases are automatically created when pushing tags (e.g., `v1.0.0`).

## License

Licensed under the MIT License.

Copyright (c) 2026 TesselStudio

See [LICENSE](LICENSE) for details.

## Acknowledgments

- Built with [Ebiten](https://ebiten.org/) game library
- Inspired by classic block-based adventure games
- Thanks to all contributors and the open-source community
