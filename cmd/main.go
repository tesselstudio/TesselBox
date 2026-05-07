package main

import (
	"fmt"

	"github.com/tesselstudio/TesselBox/pkg/ui"
)

// TesselBoxGame represents the main game state
type TesselBoxGame struct {
	fyneUI *ui.FyneKaijuBridge
}

// Version is set at build time
var Version = "dev"

func main() {
	fmt.Println("🚀 TesselBox - Fyne UI Only")
	fmt.Println("==========================")

	// Load UI configuration
	config := ui.LoadUIConfig()
	config.PrintConfig()

	// Set up environment
	config.SetupEnvironment()

	// Create and run Fyne UI
	fmt.Println("✅ Launching Fyne UI - Modern Interface")
	fmt.Println("🎮 Complete Fyne UI integration (no Kaiju UI)")

	// Create Fyne bridge
	game := &TesselBoxGame{}
	game.fyneUI = ui.NewFyneKaijuBridge(nil)

	// Show login screen
	game.fyneUI.ShowFyneLogin()

	// Run the Fyne application
	fmt.Println("🔄 Starting Fyne application...")
	game.fyneUI.Run()
}
