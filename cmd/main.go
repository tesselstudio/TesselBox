package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/tesselstudio/TesselBox/pkg/ui"
)

// TesselBoxGame represents the main game state
type TesselBoxGame struct {
	fyneUI *ui.ModernFyneUI
}

// Version is set at build time
var Version = "dev"

func main() {
	fmt.Println("🚀 TesselBox - Fyne UI Only")
	fmt.Println("==========================")

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Could not load .env file")
	}

	// Load UI configuration
	config := ui.LoadUIConfig()
	config.PrintConfig()

	// Set up environment
	config.SetupEnvironment()

	// Create and run Fyne UI
	fmt.Println("✅ Launching Fyne UI - Modern Interface")
	fmt.Println("🎮 Complete Fyne UI integration (no Kaiju UI)")

	// Create Fyne UI
	game := &TesselBoxGame{}
	game.fyneUI = ui.NewModernFyneUI()

	// Show game selection directly (no authentication)
	game.fyneUI.ShowGameSelect()

	// Run the Fyne application
	fmt.Println("🔄 Starting Fyne application...")
	game.fyneUI.Run()
}
