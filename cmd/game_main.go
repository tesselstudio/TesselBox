package main

import (
	"fmt"

	"github.com/tesselstudio/TesselBox/pkg/ui"
	"kaijuengine.com/bootstrap"
	"kaijuengine.com/engine"
)

// runGameWithFyneUI runs the actual game with Fyne UI
func runGameWithFyneUI() {
	fmt.Println("🚀 Starting TesselBox with Fyne UI...")

	// Load UI configuration
	config := ui.LoadUIConfig()
	config.PrintConfig()

	// Set up environment
	config.SetupEnvironment()

	// Check if Fyne UI should be used
	if config.UseFyneUI {
		fmt.Println("✅ Fyne UI enabled - showing modern interface")

		// Create and show Fyne login directly
		host := getEngineHost()
		if host != nil {
			// Create Fyne bridge and show login
			bridge := ui.NewFyneKaijuBridge(host.(*engine.Host), nil)
			bridge.ShowFyneLogin()

			// Run the Fyne app
			bridge.Run()
		} else {
			fmt.Println("❌ Failed to initialize engine host")
		}
	} else {
		fmt.Println("🔄 Fyne UI disabled - using bootstrap system")
		bootstrap.Main(getGame(), nil)
	}
}

// getEngineHost creates a minimal engine host for Fyne UI
func getEngineHost() interface{} {
	// For now, return nil and let the Fyne system handle it
	// In a full implementation, this would create a proper engine host
	return nil
}

// RunGameMode starts the game in actual game mode (not test mode)
func RunGameMode() {
	fmt.Println("🎮 Starting TesselBox in GAME MODE (not test mode)")

	// Load UI configuration
	config := ui.LoadUIConfig()
	config.PrintConfig()

	// Set up environment
	config.SetupEnvironment()

	if config.UseFyneUI {
		fmt.Println("✅ Launching with Fyne UI - Modern Interface")

		// Create Fyne bridge directly
		bridge := ui.NewFyneKaijuBridge(nil, nil)
		bridge.ShowFyneLogin()

		// Run the Fyne application
		bridge.Run()
	} else {
		fmt.Println("🔄 Launching with Kaiju UI")
		bootstrap.Main(getGame(), nil)
	}
}
