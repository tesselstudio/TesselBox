# TesselBox Configuration Guide

## 🎮 UI System Configuration

TesselBox now supports **two UI systems** that you can choose between:

### **Fyne UI (Modern)**
- ✅ Modern, native-looking interface
- ✅ Better visual design and user experience
- ✅ Webview-based components
- ✅ Recommended for most users

### **Kaiju UI (Original)**
- ✅ HTML-based interface
- ✅ Existing game systems
- ✅ Traditional styling
- ✅ Good for compatibility

## 🚀 Quick Start

### **Option 1: Use the Launcher Script (Recommended)**

```bash
# Run with modern Fyne UI (default)
./run.sh

# Run with original Kaiju UI
./run.sh --kaiju

# Run with hardware acceleration (Vulkan)
./run.sh --vulkan

# Run with debug mode
./run.sh --debug

# Custom window size
./run.sh --width 1280 --height 720
```

### **Option 2: Use Environment Variables**

```bash
# Enable Fyne UI (default)
export TESSELBOX_USE_FYNE=true
go run .

# Enable Kaiju UI
export TESSELBOX_USE_FYNE=false
go run .

# Enable Vulkan rendering
export TESSELBOX_USE_VULKAN=true
go run .

# Set window size
export TESSELBOX_WINDOW_WIDTH=1280
export TESSELBOX_WINDOW_HEIGHT=720
go run .

# Enable debug mode
export TESSELBOX_DEBUG=true
go run .
```

### **Option 3: Direct Go Run**

```bash
# Default configuration (Fyne UI + software rendering)
go run .

# With Vulkan hardware rendering
export VK_ICD_FILENAMES="/usr/share/vulkan/icd.d/llvmpipe_icd.x86_64.json"
go run .
```

## ⚙️ Configuration Options

| Option | Environment Variable | Default | Description |
|--------|---------------------|---------|-------------|
| UI System | `TESSELBOX_USE_FYNE` | `true` | `true` = Fyne, `false` = Kaiju |
| Rendering | `TESSELBOX_USE_VULKAN` | `false` | `true` = Vulkan, `false` = Software |
| Window Width | `TESSELBOX_WINDOW_WIDTH` | `1920` | Window width in pixels |
| Window Height | `TESSELBOX_WINDOW_HEIGHT` | `1080` | Window height in pixels |
| Debug Mode | `TESSELBOX_DEBUG` | `false` | Enable debug logging |

## 🎯 Recommended Configurations

### **For Most Users**
```bash
./run.sh
```
- Fyne UI (modern interface)
- Software rendering (best compatibility)
- 1920x1080 resolution

### **For Gaming PCs**
```bash
./run.sh --vulkan
```
- Fyne UI (modern interface)
- Vulkan rendering (hardware acceleration)
- Better performance

### **For Development**
```bash
./run.sh --debug --width 1280 --height 720
```
- Debug mode enabled
- Smaller window for development
- Detailed logging

### **For Low-End Systems**
```bash
./run.sh --kaiju --width 1024 --height 768
```
- Kaiju UI (lighter weight)
- Software rendering
- Smaller resolution

## 🔧 Troubleshooting

### **Game Doesn't Start**
```bash
# Try software rendering
export VK_ICD_FILENAMES="/usr/share/vulkan/icd.d/llvmpipe_icd.x86_64.json"
go run .
```

### **UI Looks Bad**
```bash
# Try the other UI system
./run.sh --kaiju  # or --fyne
```

### **Performance Issues**
```bash
# Try software rendering
./run.sh --software

# Or lower resolution
./run.sh --width 1280 --height 720
```

### **Debug Mode**
```bash
# Enable debug to see what's happening
./run.sh --debug
```

## 🎨 UI Features

### **Fyne UI Features**
- ✅ Modern login screen
- ✅ Game mode selection
- ✅ Better visual design
- ✅ Native-looking components
- ✅ Smooth transitions

### **Kaiju UI Features**
- ✅ HTML-based interface
- ✅ Existing game systems
- ✅ Custom styling
- ✅ Proven stability

## 🔄 Switching Between UI Systems

You can switch between UI systems at any time:

```bash
# Switch to Fyne
export TESSELBOX_USE_FYNE=true
go run .

# Switch to Kaiju
export TESSELBOX_USE_FYNE=false
go run .
```

The game will automatically use the configured UI system on startup.

## 📱 Mobile Support

For mobile devices or tablets:

```bash
# Smaller window size
./run.sh --width 800 --height 600

# Or portrait mode
./run.sh --width 600 --height 800
```

## 🎮 Next Steps

1. **Try the default**: `./run.sh`
2. **Experiment with options**: `./run.sh --help`
3. **Find your preferred configuration**
4. **Enjoy the modern Fyne interface!**

---

**Note**: The Fyne UI system is the recommended choice for the best user experience. It provides a modern, polished interface while maintaining full compatibility with the existing game systems.
