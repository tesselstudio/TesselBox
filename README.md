# TesselBox 3D

A next-generation 3D hexagonal block-based adventure game built with the Kaiju Engine, featuring exploration, crafting, combat, survival elements, and multiplayer support.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8.svg)](https://golang.org/)

## Overview

TesselBox 3D is a cutting-edge rewrite of the original TesselBox game, now featuring true 3D hexagonal prism blocks and advanced rendering powered by the Kaiju Engine with Vulkan. Explore vast 3D worlds, gather resources, craft tools, build complex structures, survive environmental challenges, and play with friends in multiplayer mode.

## Key Features

### 🎮 Core Gameplay
- **3D Hexagonal Prism System** - Revolutionary 6-sided prism blocks with 6-orientation rotation
- **Advanced Rendering** - Vulkan-powered graphics with modern lighting and effects
- **Procedural World Generation** - Infinite 3D worlds with varied biomes and terrain
- **Complex Block System** - 11+ block types with attachment mechanics and rotation
- **Hexagonal Coordinate System** - Mathematical precision in 3D space

### 🛠️ Crafting & Building
- **Comprehensive Crafting System** - 15+ recipes with shaped and shapeless crafting
- **Advanced Inventory Management** - Stacking, slot management, and hotbar system
- **Rich Item System** - 20+ items including tools, weapons, armor, and materials
- **Block Placement Mechanics** - Advanced attachment validation and collision detection

### 🌍 Survival & Environment
- **Health & Hunger Systems** - Complex survival mechanics with regeneration and exhaustion
- **Environmental System** - Day/night cycle, weather, seasons, and biomes
- **Temperature Regulation** - Body temperature mechanics with environmental effects
- **Environmental Hazards** - Fire, water, lava, poison, radiation, cold, and heat
- **Experience & Leveling** - Character progression with skill points

### 🌐 Multiplayer
- **Advanced Networking** - TCP-based protocol with message encoding/decoding
- **Server Architecture** - Client management, world synchronization, and connection handling
- **Client-Side Prediction** - Smooth gameplay with reconciliation
- **Real-time Updates** - Player movement, block placement, and world state synchronization

### 🎨 User Interface
- **Modern UI System** - Complete component system with panels, buttons, labels, progress bars
- **Comprehensive HUD** - Health/hunger/stamina bars, hotbar, crosshair, and debug info
- **Event Management** - User interaction handling and layout systems
- **Inventory UI Integration** - Seamless inventory management interface

## Technical Architecture

### Engine & Technology
- **Kaiju Engine** - Custom Go game engine with Vulkan backend
- **Hexagonal Prism Rendering** - 24-vertex mesh generation with proper normals and UVs
- **Custom Math Library** - Matrix operations for 3D transformations
- **Modular Design** - Clean separation between systems for maintainability

### Core Systems
- **Block System** (`pkg/blocks/`) - Hexagonal prism blocks with attachment mechanics
- **World Generation** (`pkg/world/`) - Hexagonal coordinate system and terrain generation
- **Networking** (`pkg/network/`) - TCP-based multiplayer with client-side prediction
- **Crafting** (`pkg/crafting/`) - Item system, inventory management, and recipes
- **Survival** (`pkg/survival/`) - Health, hunger, stamina, and environmental systems
- **UI System** (`pkg/ui/`) - Modern component-based user interface

## Building

### Requirements

- Go 1.25 or later
- Vulkan-compatible graphics driver
- Kaiju Engine (included as submodule)

### Quick Build

```bash
# Clone the repository
git clone https://github.com/tesselstudio/TesselBox.git
cd TesselBox

# Build the game
go build -o tesselbox ./cmd/main.go
```

### Development Build

```bash
# Ensure dependencies are available
go mod tidy

# Build with debug information
go build -gcflags="all=-N -l" -o tesselbox-debug ./cmd/main.go
```

## Running

```bash
# Run the game
./tesselbox

# Run with debug mode
./tesselbox --debug
```

## Project Structure

```
TesselBox/
├── cmd/
│   └── main.go              # Main entry point using Kaiju Engine
├── pkg/
│   ├── blocks/             # Hexagonal prism block system
│   │   ├── geometry.go     # 3D mesh generation
│   │   ├── registry.go     # Block type definitions
│   │   └── placement.go    # Block placement mechanics
│   ├── crafting/           # Crafting and inventory system
│   │   ├── items.go        # Item definitions and registry
│   │   ├── inventory.go    # Inventory management
│   │   └── recipes.go      # Crafting recipes
│   ├── network/            # Multiplayer networking
│   │   ├── protocol.go     # Message encoding/decoding
│   │   ├── server.go       # Server architecture
│   │   └── client.go       # Client networking
│   ├── survival/           # Survival mechanics
│   │   ├── health.go       # Health, hunger, stamina systems
│   │   └── environment.go  # Environmental systems
│   ├── ui/                 # User interface system
│   │   ├── components.go   # UI components
│   │   └── hud.go          # Heads-up display
│   └── world/              # World generation
│       └── coordinates.go   # Hexagonal coordinate system
├── kaiju/                  # Kaiju Engine source (submodule)
├── config/                 # Configuration files
├── scripts/                # Build and utility scripts
└── go.mod                  # Go module definition
```

## Development Status

| Component | Status | Notes |
|-----------|--------|-------|
| Core Engine | ✅ Complete | Kaiju Engine integration |
| 3D Rendering | ✅ Complete | Hexagonal prism meshes |
| Block System | ✅ Complete | 11+ block types with rotation |
| World Generation | ✅ Complete | Hexagonal coordinate system |
| Multiplayer | ✅ Complete | TCP networking with prediction |
| Crafting | ✅ Complete | Items, inventory, recipes |
| Survival | ✅ Complete | Health, hunger, environment |
| UI System | ✅ Complete | Components and HUD |
| Save/Load | 🚧 In Progress | World persistence |
| Menu System | 🚧 In Progress | Main navigation |

## Documentation

- [Contributing Guidelines](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Security Policy](SECURITY.md)

## Community

- [Discussions](https://github.com/tesselstudio/TesselBox/discussions) - Ask questions, share ideas
- [Issues](https://github.com/tesselstudio/TesselBox/issues) - Report bugs, request features

## License

Licensed under the MIT License.

Copyright (c) 2026 TesselStudio

See [LICENSE](LICENSE) for details.

## Acknowledgments

- Built with [Kaiju Engine](https://kaijuengine.com/) - Vulkan-powered Go game engine
- Inspired by classic block-based adventure games
- Thanks to all contributors and the open-source community

## Architecture Highlights

### Revolutionary Hexagonal System
TesselBox 3D introduces a true 3D hexagonal prism system that offers unprecedented creative flexibility compared to traditional cubic block games:

- **6-sided prisms** instead of cubes for more organic building
- **6-orientation rotation** for complex architectural designs
- **Attachment mechanics** with face-based validation
- **Mathematical precision** with axial coordinate systems

### Advanced Multiplayer
- **Client-side prediction** for smooth gameplay
- **TCP-based protocol** with message encoding/decoding
- **World state synchronization** across all clients
- **Scalable server architecture** supporting multiple players

### Modern Game Systems
- **Complex survival mechanics** with health, hunger, stamina, and environmental factors
- **Rich crafting system** with shaped and shapeless recipes
- **Advanced inventory management** with stacking and hotbar systems
- **Modern UI framework** with component-based design

This implementation represents a significant advancement in voxel-based gaming, combining the creative freedom of block-based games with the precision of modern 3D graphics and the social aspects of multiplayer gaming.
