# TesselBox

A hexagonal block-based adventure game with exploration, crafting, combat, and survival elements.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview

TesselBox is an open-source game built with a unique hexagonal grid system. Explore vast worlds, gather resources, craft tools, build structures, and survive against hostile mobs.

## Repository Structure

TesselBox is organized into 5 separate repositories for modularity:

| Repository | Description | Link |
|------------|-------------|------|
| **TesselBox-main** | Main documentation and discussion hub | You're here! |
| **TesselBox-pc** | PC/Desktop version source code | [GitHub](https://github.com/tesselstudio/TesselBox-pc) |
| **TesselBox-mobile** | Mobile (Android/iOS) version source code | [GitHub](https://github.com/tesselstudio/TesselBox-mobile) |
| **TesselBox-assets** | Shared game assets and configuration | [GitHub](https://github.com/tesselstudio/TesselBox-assets) |
| **TesselBox-build** | Unified build scripts and CI/CD | [GitHub](https://github.com/tesselstudio/TesselBox-build) |

## Features

- **Hexagonal Grid System** - Unique 6-sided block world
- **Procedural Generation** - Infinite worlds with varied biomes
- **Crafting System** - Craft tools, weapons, and building materials
- **Survival Mode** - Manage health, hunger, and fend off enemies
- **Creative Mode** - Build without limits
- **Dimensions** - Travel between overworld and Randomland
- **Multi-platform** - Play on PC (Windows/Linux/macOS) or Mobile (Android/iOS)

## Getting Started

### Quick Start

```bash
# Clone all repositories
git clone https://github.com/tesselstudio/TesselBox-pc.git
git clone https://github.com/tesselstudio/TesselBox-mobile.git
git clone https://github.com/tesselstudio/TesselBox-assets.git
git clone https://github.com/tesselstudio/TesselBox-build.git

# Build using the unified build system
cd TesselBox-build
make pc        # Build PC version
make mobile    # Build Mobile version
```

### Platform-Specific

- **[PC Build](https://github.com/tesselstudio/TesselBox-pc)** - Keyboard and mouse controls
- **[Mobile Build](https://github.com/tesselstudio/TesselBox-mobile)** - Touch gesture controls

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

## Documentation

- [Architecture Overview](docs/ARCHITECTURE.md)
- [Contributing Guidelines](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Security Policy](SECURITY.md)

## Community

- [Discussions](https://github.com/tesselstudio/TesselBox-main/discussions) - Ask questions, share ideas
- [Issues](https://github.com/tesselstudio/TesselBox-main/issues) - Report bugs, request features
- [Wiki](https://github.com/tesselstudio/TesselBox-main/wiki) - Community documentation

## Development Status

| Platform | Status |
|----------|--------|
| PC (Windows/Linux/macOS) | Active |
| Mobile (Android/iOS) | Active |

## License

All TesselBox repositories are licensed under the MIT License.

Copyright (c) 2026 TesselStudio

See [LICENSE](LICENSE) for details.

## Acknowledgments

- Built with [Ebiten](https://ebiten.org/) game library
- Inspired by classic block-based adventure games
- Thanks to all contributors and the open-source community
