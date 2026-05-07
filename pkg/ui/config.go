package ui

import (
	"os"
	"strconv"
)

// UIConfig holds configuration for UI systems
type UIConfig struct {
	UseFyneUI    bool
	UseVulkan    bool
	WindowWidth  int
	WindowHeight int
	DebugMode    bool
}

// DefaultUIConfig returns default UI configuration
func DefaultUIConfig() *UIConfig {
	return &UIConfig{
		UseFyneUI:    true,  // Enable Fyne UI by default
		UseVulkan:    false, // Start with software rendering
		WindowWidth:  1920,
		WindowHeight: 1080,
		DebugMode:    false,
	}
}

// LoadUIConfig loads UI configuration from environment variables
func LoadUIConfig() *UIConfig {
	config := DefaultUIConfig()

	// Check for Fyne UI preference
	if useFyne := os.Getenv("TESSELBOX_USE_FYNE"); useFyne != "" {
		if parsed, err := strconv.ParseBool(useFyne); err == nil {
			config.UseFyneUI = parsed
		}
	}

	// Check for Vulkan rendering
	if useVulkan := os.Getenv("TESSELBOX_USE_VULKAN"); useVulkan != "" {
		if parsed, err := strconv.ParseBool(useVulkan); err == nil {
			config.UseVulkan = parsed
		}
	}

	// Check for window dimensions
	if width := os.Getenv("TESSELBOX_WINDOW_WIDTH"); width != "" {
		if parsed, err := strconv.Atoi(width); err == nil && parsed > 0 {
			config.WindowWidth = parsed
		}
	}

	if height := os.Getenv("TESSELBOX_WINDOW_HEIGHT"); height != "" {
		if parsed, err := strconv.Atoi(height); err == nil && parsed > 0 {
			config.WindowHeight = parsed
		}
	}

	// Check for debug mode
	if debug := os.Getenv("TESSELBOX_DEBUG"); debug != "" {
		if parsed, err := strconv.ParseBool(debug); err == nil {
			config.DebugMode = parsed
		}
	}

	return config
}

// PrintConfig prints the current UI configuration
func (c *UIConfig) PrintConfig() {
	println("=== TesselBox UI Configuration ===")
	println("UI System:", c.getUITypeName())
	println("Rendering:", c.getRenderingTypeName())
	println("Window Size:", c.WindowWidth, "x", c.WindowHeight)
	println("Debug Mode:", c.DebugMode)
	println("=====================================")
}

// getUITypeName returns the name of the UI system being used
func (c *UIConfig) getUITypeName() string {
	if c.UseFyneUI {
		return "Fyne (Modern)"
	}
	return "Kaiju (HTML-based)"
}

// getRenderingTypeName returns the name of the rendering system
func (c *UIConfig) getRenderingTypeName() string {
	if c.UseVulkan {
		return "Vulkan (Hardware)"
	}
	return "Software (CPU)"
}

// SetupEnvironment sets up environment variables for optimal performance
func (c *UIConfig) SetupEnvironment() {
	if !c.UseVulkan {
		// Set up software rendering for better compatibility
		os.Setenv("VK_ICD_FILENAMES", "/usr/share/vulkan/icd.d/llvmpipe_icd.x86_64.json")
	}

	if c.DebugMode {
		os.Setenv("RUST_LOG", "debug")
		os.Setenv("FYNE_DEBUG", "1")
	}
}

// GetCommandLineArgs returns command line arguments for the game
func (c *UIConfig) GetCommandLineArgs() []string {
	var args []string

	if c.UseVulkan {
		args = append(args, "--vulkan")
	}

	if c.DebugMode {
		args = append(args, "--debug")
	}

	if c.WindowWidth != 1920 || c.WindowHeight != 1080 {
		args = append(args, "--window-size", strconv.Itoa(c.WindowWidth), strconv.Itoa(c.WindowHeight))
	}

	return args
}
