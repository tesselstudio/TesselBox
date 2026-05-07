package main

import (
	"fmt"
	"os"

	"github.com/tesselstudio/TesselBox/pkg/ui"
)

func main() {
	fmt.Println("🚀 TesselBox Fyne UI Launcher")
	fmt.Println("================================")

	// Load UI configuration
	config := ui.LoadUIConfig()
	config.PrintConfig()

	// Set up environment
	config.SetupEnvironment()

	if config.UseFyneUI {
		fmt.Println("✅ Launching Fyne UI - Modern Interface")
		fmt.Println("🎮 This is the REAL game interface (not test mode)")
		fmt.Println("")

		// Create Fyne bridge directly without engine host
		// This will show the modern login screen
		bridge := ui.NewFyneKaijuBridge(nil)
		bridge.ShowFyneLogin()

		// Run the Fyne application
		fmt.Println("🔄 Starting Fyne application...")
		bridge.Run()

	} else {
		fmt.Println("❌ Fyne UI is disabled")
		fmt.Println("💡 Set TESSELBOX_USE_FYNE=true to enable Fyne UI")
		os.Exit(1)
	}
}
