# TesselBox

A browser-based voxel sandbox game built with Go and WebGL. Explore, build, and survive in a block-based world directly in your browser.

![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg?url=https://github.com/tesselstudio/TesselBox)
![License](https://img.shields.io/badge/license-MIT-green.svg?url=https://github.com/tesselstudio/TesselBox)
<img src="https://komarev.com/ghpvc/?username=tesselstudio&color=58A6FF&style=flat-square&label=Views" alt="Views"/>
![GitHub Issues](https://img.shields.io/github/issues/tesselstudio/TesselBox.svg?url=https://github.com/tesselstudio/TesselBox)
![GitHub Pull Request](https://img.shields.io/github/issues-pr/tesselstudio/TesselBox.svg?url=https://github.com/tesselstudio/TesselBox)
![GitHub Contributors](https://img.shields.io/github/contributors/tesselstudio/TesselBox.svg?url=https://github.com/tesselstudio/TesselBox)
![GitHub Last Commit](https://img.shields.io/github/last-commit/tesselstudio/TesselBox.svg?url=https://github.com/tesselstudio/TesselBox)
![WebAssembly](https://img.shields.io/badge/WebAssembly-Enabled-blue.svg?url=https://github.com/tesselstudio/TesselBox)
![GitHub Repo Size](https://img.shields.io/github/repo-size/tesselstudio/TesselBox.svg?url=https://github.com/tesselstudio/TesselBox)

## Quick Start


```bash
# Build web version
make build-web

# Serve locally
make run-web
# Opens http://localhost:8080
```

## System Requirements

- **Browser**: Chrome 79+, Firefox 72+, Safari 15+, Edge 79+
- **Graphics**: WebGL 2.0 support (most modern browsers)
- **RAM**: 4GB minimum
- **Storage**: 100MB

## Controls

|Key|Action|
|-----|--------|
|WASD|Move (works in all camera modes)|
|Space|Jump|
|Mouse|Look around|
|Left Click|Break block|
|Right Click|Place block|
|E|Open inventory|
|C|Open crafting|
|Right Ctrl/C|Cycle camera perspective (1st → 2nd → 3rd person)|
|Q|Drop item|
|ESC|Pause menu|
|F3|Toggle debug info|
|1-9|Select hotbar slot|

## Project Structure

```
TesselBox/
├── cmd/           # Entry points (web)
├── pkg/           # Core packages
│   ├── audio/     # Audio system
│   ├── blocks/    # Block system
│   ├── config/    # Configuration
│   ├── content/   # Content database
│   ├── crafting/  # Crafting system
│   ├── effects/   # Particle effects
│   ├── game/      # Game logic & state
│   ├── logger/    # Logging system
│   ├── network/   # Multiplayer networking
│   ├── player/    # Player & inventory
│   ├── survival/  # Survival mechanics
│   ├── types/     # Type definitions
│   ├── webgl/     # WebGL rendering
│   └── world/     # World generation & saving
├── web/           # Web frontend files
└── dist/          # Build outputs
```

## Architecture

TesselBox runs entirely in the browser using WebAssembly and WebGL:

- **Go → WebAssembly**: Game logic compiled to WASM for browser execution
- **WebGL Rendering**: Hardware-accelerated graphics in the browser
- **Custom UI System**: Built with pure WebGL/HTML for game interfaces
- **ECS Architecture**: Entity Component System for game organization

## Features

- **Infinite Generated Worlds** - Procedurally generated worlds with persistent saves
- **Block Manipulation** - Place, break, and interact with blocks to build structures
- **Inventory Management** - Collect items and manage your inventory
- **Crafting System** - Combine materials to craft tools and items
- **Multiplayer Support** - Host and join multiplayer servers
- **No Installation Required** - Play directly in your browser
- **Positional Audio** - Immersive audio with music and sound effects

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

```bash
# Build web version
make build-web

# Run locally
make run-web

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

TesselBox is a voxel-based sandbox game built with Go and WebGL. The project began in October 2021 as a side project. The initial commit was made on October 25, 2021. The project has been actively developed since then with regular commits and major updates, including a complete migration to browser-based WebAssembly in 2026.

## Credits

Thanks to all the contributors who have helped make TesselBox possible.

---

Built with Go and WebGL for the browser.