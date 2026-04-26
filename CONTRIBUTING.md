# Contributing to TesselBox

Thank you for your interest in contributing to TesselBox! This document provides guidelines for contributing to the project.

## How to Contribute

### Reporting Bugs

- Use the [issue tracker](https://github.com/tesselstudio/TesselBox-main/issues) on the main repository
- Describe the bug in detail
- Include steps to reproduce
- Specify your platform (PC/Mobile, OS version)
- Include any error messages or logs

### Suggesting Features

- Open a discussion in the [Discussions](https://github.com/tesselstudio/TesselBox-main/discussions) section
- Describe the feature and its use case
- Be open to feedback from maintainers

### Code Contributions

1. **Fork the appropriate repository**:
   - `TesselBox-pc` for desktop changes
   - `TesselBox-mobile` for mobile changes
   - `TesselBox-assets` for game content changes
   - `TesselBox-build` for build system changes

2. **Create a branch**: `git checkout -b feature/your-feature-name`

3. **Make your changes** with clear, documented code

4. **Test your changes**:
   ```bash
   make test  # Run tests
   make pc    # Build PC version
   ```

5. **Commit with clear messages**:
   ```
   feat: add new block type
   fix: resolve rendering issue on mobile
   docs: update build instructions
   ```

6. **Push and create a Pull Request**

## Development Setup

```bash
# Clone all repos
git clone https://github.com/tesselstudio/TesselBox-pc.git
git clone https://github.com/tesselstudio/TesselBox-mobile.git
git clone https://github.com/tesselstudio/TesselBox-assets.git
git clone https://github.com/tesselstudio/TesselBox-build.git

# Build
cd TesselBox-build
make pc
```

## Code Standards

- Follow existing code style
- Add comments for complex logic
- Write tests for new features
- Ensure cross-platform compatibility

## Areas for Contribution

- **Gameplay**: New blocks, items, mobs
- **UI/UX**: Interface improvements, accessibility
- **Performance**: Optimization, rendering improvements
- **Documentation**: Tutorials, API docs
- **Platform Support**: Platform-specific features

## Questions?

Join our [Discussions](https://github.com/tesselstudio/TesselBox-main/discussions) or open an issue!
