# TesselBox

A voxel-based sandbox game built with Go and the Kaiju Engine. Explore, build, and survive in a block-based world.

![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg?url=https://github.com/tesselstudio/TesselBox)
![License](https://img.shields.io/badge/license-MIT-green.svg?url=https://github.com/tesselstudio/TesselBox)
<img src="https://komarev.com/ghpvc/?username=tesselstudio&color=58A6FF&style=flat-square&label=Views" alt="Views"/>
![GitHub Issues](https://img.shields.io/github/issues/tesselstudio/TesselBox.svg?url=https://github.com/tesselstudio/TesselBox)
![GitHub Pull Request](https://img.shields.io/github/issues-pr/tesselstudio/TesselBox.svg?url=https://github.com/tesselstudio/TesselBox)
![GitHub Contributors](https://img.shields.io/github/contributors/tesselstudio/TesselBox.svg?url=https://github.com/tesselstudio/TesselBox)
![GitHub Last Commit](https://img.shields.io/github/last-commit/tesselstudio/TesselBox.svg?url=https://github.com/tesselstudio/TesselBox)
![Kaiju Engine](https://img.shields.io/badge/Kaiju-Engine-blue.svg?url=https://github.com/tesselstudio/TesselBox)
![Version](https://img.shields.io/badge/Version-0.1.0-blue.svg?url=https://github.com/tesselstudio/TesselBox)
![GitHub Repo Size](https://img.shields.io/github/repo-size/tesselstudio/TesselBox.svg?url=https://github.com/tesselstudio/TesselBox)


## Quick Start


## System Requirements

TesselBox requires the following to run:

<details>
<summary>Linux</summary>

- **OS**: Ubuntu 20.04+ or any other Linux distribution that supports Vulkan
- **Graphics**: Vulkan-compatible GPU
- **RAM**: 4GB minimum, 8GB recommended
- **Storage**: 500MB

</details>

<details>
<summary>Windows</summary>

- **OS**: Windows 10+
- **Graphics**: Vulkan-compatible GPU
- **RAM**: 4GB minimum, 8GB recommended
- **Storage**: 500MB

</details>

<details>
<summary>macOS</summary>

- **OS**: macOS 10.15+
- **Graphics**: Vulkan-compatible GPU
- **RAM**: 4GB minimum, 8GB recommended
- **Storage**: 500MB

</details>

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- Make

### Run

```bash
make run
```

### Build

```bash
# Development build
make dev

# Linux
make build-linux

# Windows
make build-windows

# macOS
make build-macos

# All platforms
make build-all
```

## Controls

|Key|Action|
|-----|--------|
|WASD|Move|
|Space|Jump|
|Mouse|Look around|
|Left Click|Break block|
|Right Click|Place block|
|E|Open inventory|
|C|Open crafting|
|Q|Drop item|
|ESC|Pause menu|
|F3|Toggle debug info|
|1-9|Select hotbar slot|

## Project Structure

```text
TesselBox/
├── cmd/           # Main entry point
├── pkg/           # Core packages
│   ├── audio/     # Audio system
│   ├── crafting/  # Crafting system
│   ├── game/      # Game logic & state
│   ├── network/   # Multiplayer networking
│   ├── world/     # World generation & saving
│   └── player/    # Player & inventory
├── game_content/  # Assets & UI files
├── kaiju/         # Kaiju Engine submodule
├── scripts/       # Build scripts
└── dist/          # Build outputs
```

## Architecture

TesselBox is built on the [Kaiju Engine](https://github.com/KaijuEngine), providing:

- Entity Component System (ECS)
- UI/Markup system
- Asset management
- Input handling
- Rendering pipeline
- More details can be found in the [Kaiju Engine documentation](https://github.com/KaijuEngine)


## Features

- **Infinite Generated Worlds** - Procedurally generated worlds with persistent saves
- **Block Manipulation** - Place, break, and interact with blocks to build structures
- **Inventory Management** - Collect items and manage your inventory
- **Crafting System** - Combine materials to craft tools and items
- **Multiplayer Support** - Host and join multiplayer servers
- **Cross-Platform Compatibility** - Runs on Linux, Windows, macOS, Android, and iOS
- **Positional Audio** - Audio with music and sound effects that is positional

## Multiplayer

### Host a Server

```bash
make run-server
# or
./tesselbox --server
```

Default port: `8080`

### Connect to a Server

1. Open Multiplayer from the main menu
2. Enter server address (e.g., `localhost:8080`)
3. Click Connect

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

```bash
# Run tests
make test

# Format code
make fmt

# Lint
make lint

# Download dependencies
make deps
```

## License

MIT License - see [LICENSE](LICENSE) file.

## Security

Report security issues privately - see [SECURITY.md](SECURITY.md).


## History

TesselBox is a game built on the Kaiju Engine. The project began in October 2021 as a side project. The initial commit was made on October 25, 2021. The project has been actively developed since then with regular commits and major updates.

## Development

The development journey of TesselBox is ongoing. The project has undergone several major updates, including the addition of multiplayer support, improved graphics, and cross-platform support. The project is actively maintained and new features are being added regularly.

## Credits

Thanks to all the contributors and the Kaiju Engine team for their amazing work.

---

Built with Go and the Kaiju Engine.
