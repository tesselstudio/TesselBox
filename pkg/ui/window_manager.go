package ui

import (
	"fmt"
	"sync"

	"github.com/tesselstudio/TesselBox/pkg/game"
	"github.com/tesselstudio/TesselBox/pkg/opengl"
)

// WindowManager manages switching between Fyne UI and OpenGL windows
type WindowManager struct {
	mu sync.RWMutex

	// Fyne window
	fyneUI *ModernFyneUI

	// Game context
	gameController *game.Controller
	renderEngine   *opengl.Engine
	gameLoop       *game.GameLoop
	inputHandler   *game.InputHandler

	// State
	isGameRunning bool
}

// NewWindowManager creates a new window manager
func NewWindowManager(fyneUI *ModernFyneUI) *WindowManager {
	return &WindowManager{
		fyneUI: fyneUI,
	}
}

// StartGame initializes and starts the game
func (wm *WindowManager) StartGame(worldName string, seed int64, isMultiplayer bool) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.isGameRunning {
		return fmt.Errorf("game already running")
	}

	// Hide Fyne UI
	if wm.fyneUI != nil {
		wm.fyneUI.Hide()
	}

	// Create OpenGL engine
	engine, err := opengl.NewEngine(1280, 720, "TesselBox - Game")
	if err != nil {
		return fmt.Errorf("failed to create OpenGL engine: %v", err)
	}

	// Create game controller
	controller := game.NewController()

	// Create input handler
	window := engine.GetWindow()
	inputHandler := game.NewInputHandler(window)

	// Create game loop
	gameLoop := game.NewGameLoop(controller, engine, 60)

	// Store references
	wm.renderEngine = engine
	wm.gameController = controller
	wm.inputHandler = inputHandler
	wm.gameLoop = gameLoop
	wm.isGameRunning = true

	// Lock cursor to window
	inputHandler.LockMouse()

	fmt.Printf("🎮 Starting %s world: '%s' (seed: %d)\n",
		map[bool]string{true: "Multiplayer", false: "Singleplayer"}[isMultiplayer],
		worldName,
		seed)

	// Start world
	controller.StartWorld(worldName, seed)

	// Run game loop in goroutine
	go func() {
		defer wm.cleanupGame()
		if err := gameLoop.Run(); err != nil {
			fmt.Printf("❌ Game loop error: %v\n", err)
		}
	}()

	return nil
}

// ProcessInput processes input for the currently running game
func (wm *WindowManager) ProcessInput() {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if wm.inputHandler == nil || wm.gameController == nil {
		return
	}

	// Process input into controller
	wm.inputHandler.ProcessInput(wm.gameController)

	// Reset mouse delta after frame
	defer wm.inputHandler.ResetMouseDelta()
}

// cleanupGame cleans up game resources
func (wm *WindowManager) cleanupGame() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.gameController != nil {
		wm.gameController.Stop()
	}

	if wm.inputHandler != nil {
		wm.inputHandler.UnlockMouse()
	}

	if wm.renderEngine != nil {
		wm.renderEngine.Cleanup()
	}

	wm.isGameRunning = false

	// Show Fyne UI again
	if wm.fyneUI != nil {
		wm.fyneUI.Show()
	}

	fmt.Println("🎮 Game stopped")
}

// IsGameRunning returns whether a game is currently running
func (wm *WindowManager) IsGameRunning() bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.isGameRunning
}

// GetGameController returns the current game controller
func (wm *WindowManager) GetGameController() *game.Controller {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.gameController
}

// GetRenderEngine returns the current render engine
func (wm *WindowManager) GetRenderEngine() *opengl.Engine {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.renderEngine
}

// CreateWorldDialog shows a dialog for creating a new world
func (wm *WindowManager) CreateWorldDialog() (worldName string, seed int64, cancelled bool) {
	// This would be implemented with Fyne dialogs
	// For now, return defaults
	return "New World", 12345, false
}

// SelectWorldDialog shows a dialog for selecting an existing world
func (wm *WindowManager) SelectWorldDialog() (worldName string, cancelled bool) {
	// This would be implemented with Fyne dialogs
	// For now, return defaults
	return "World 1", false
}

// SelectServerDialog shows a dialog for selecting a multiplayer server
func (wm *WindowManager) SelectServerDialog() (address string, username string, cancelled bool) {
	// This would be implemented with Fyne dialogs
	// For now, return defaults
	return "localhost:25565", "Player", false
}
