package main

import (
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/tesselstudio/TesselBox/pkg/opengl"
	"github.com/tesselstudio/TesselBox/pkg/opengl_ui"
)

// TesselBoxGame represents the main game state
type TesselBoxGame struct {
	engine *opengl.Engine
	menu   *opengl_ui.OpenGLMenu
}

// Version is set at build time
var Version = "dev"

func main() {
	fmt.Println("🚀 TesselBox - OpenGL UI Only")
	fmt.Println("==========================")

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Could not load .env file")
	}

	// Ensure GLFW is properly terminated when the application exits
	defer func() {
		fmt.Println("🧹 Cleaning up GLFW resources...")
		opengl.TerminateGLFW()
	}()

	// Main loop: run menu -> optionally run game -> repeat
	for {
		// Create OpenGL engine for menu
		fmt.Println("✅ Creating OpenGL engine for menu...")
		engine, err := opengl.NewEngine(1280, 720, "TesselBox - Main Menu")
		if err != nil {
			log.Fatalf("Failed to create OpenGL engine: %v", err)
		}

		game := &TesselBoxGame{
			engine: engine,
		}

		// Create OpenGL menu
		fmt.Println("✅ Creating OpenGL menu system...")
		game.menu, err = opengl_ui.NewOpenGLMenu(engine)
		if err != nil {
			log.Fatalf("Failed to create OpenGL menu: %v", err)
		}

		// Run menu loop
		fmt.Println("🔄 Starting OpenGL menu loop...")
		game.runMenuLoop()

		// Check if user wants to play the game
		if game.menu.ShouldRunGame() {
			fmt.Println("🎮 Starting game...")
			game.runGameLoop()
			fmt.Println("🔄 Game session ended, restarting menu...")
			// Loop continues, creating new menu
		} else {
			// User closed menu without starting game
			fmt.Println("👋 Goodbye!")
			break
		}

		// Cleanup current engine
		if game.engine != nil {
			game.engine.Cleanup()
		}
	}
}

// runMenuLoop runs the main menu loop
func (g *TesselBoxGame) runMenuLoop() {
	// Set up menu update timer
	lastTime := time.Now()

	for !g.engine.ShouldClose() {
		currentTime := time.Now()
		_ = currentTime.Sub(lastTime).Seconds()
		lastTime = currentTime

		// Handle input
		g.engine.PollEvents()

		// Update menu
		g.menu.Update()

		// Begin frame (clears screen)
		g.engine.BeginFrame()

		// Render menu
		g.menu.Render()

		// End frame (swaps buffers)
		g.engine.EndFrame()

		// Check if game should start
		if g.menu.ShouldRunGame() && !g.menu.IsGameRunning() {
			break
		}

		// Frame rate limiting
		time.Sleep(16 * time.Millisecond)
	}
}

// runGameLoop runs the game when started from menu
func (g *TesselBoxGame) runGameLoop() {
	// Game is already started by the menu system
	// We just need to wait for it to complete
	for g.menu.IsGameRunning() {
		g.engine.PollEvents()
		time.Sleep(16 * time.Millisecond) // ~60 FPS
	}
}
