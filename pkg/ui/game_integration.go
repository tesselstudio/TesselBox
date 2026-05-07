package ui

import (
	"kaijuengine.com/engine"
)

// GameUIIntegration handles UI integration with the main game
type GameUIIntegration struct {
	config      *UIConfig
	menuManager *MenuManager
	host        *engine.Host
	fyneBridge  *FyneKaijuBridge
}

// NewGameUIIntegration creates a new game UI integration
func NewGameUIIntegration(host *engine.Host) *GameUIIntegration {
	config := LoadUIConfig()
	config.PrintConfig()
	config.SetupEnvironment()
	
	integration := &GameUIIntegration{
		config: config,
		host:   host,
	}
	
	return integration
}

// Initialize sets up the UI system based on configuration
func (g *GameUIIntegration) Initialize() {
	// Create menu manager
	screenWidth := float32(g.config.WindowWidth)
	screenHeight := float32(g.config.WindowHeight)
	g.menuManager = NewMenuManager(screenWidth, screenHeight)
	
	// Enable Fyne UI if configured
	if g.config.UseFyneUI {
		g.menuManager.EnableFyneUI(g.host)
		g.fyneBridge = g.menuManager.fyneBridge
		println("✅ Fyne UI enabled")
	} else {
		println("✅ Kaiju UI enabled")
	}
	
	// Set up game state transitions
	g.setupGameTransitions()
}

// setupGameTransitions sets up transitions between game states
func (g *GameUIIntegration) setupGameTransitions() {
	// Set up menu callbacks
	g.menuManager.SetOnStartGame(func(worldName string, seed int64) {
		println("🎮 Starting game:", worldName)
		// TODO: Transition to gameplay
	})
	
	g.menuManager.SetOnResumeGame(func() {
		println("🔄 Resuming game")
		// TODO: Resume gameplay
	})
	
	g.menuManager.SetOnQuitToDesktop(func() {
		println("👋 Quitting to desktop")
		if g.host != nil {
			g.host.Close()
		}
	})
}

// ShowLogin displays the login screen
func (g *GameUIIntegration) ShowLogin() {
	if g.config.UseFyneUI && g.fyneBridge != nil {
		g.fyneBridge.ShowFyneLogin()
	} else {
		g.menuManager.ShowLoginScreen()
	}
}

// ShowMainMenu displays the main menu
func (g *GameUIIntegration) ShowMainMenu() {
	g.menuManager.ShowMenu(MenuTypeMain)
}

// GetMenuManager returns the menu manager
func (g *GameUIIntegration) GetMenuManager() *MenuManager {
	return g.menuManager
}

// GetConfig returns the UI configuration
func (g *GameUIIntegration) GetConfig() *UIConfig {
	return g.config
}

// IsFyneActive returns true if Fyne UI is currently active
func (g *GameUIIntegration) IsFyneActive() bool {
	return g.menuManager.IsFyneActive()
}

// SwitchToKaiju switches to Kaiju UI system
func (g *GameUIIntegration) SwitchToKaiju() {
	if g.fyneBridge != nil {
		g.fyneBridge.SwitchToKaiju()
	}
}

// SwitchToFyne switches to Fyne UI system
func (g *GameUIIntegration) SwitchToFyne() {
	if g.fyneBridge != nil {
		g.fyneBridge.ShowFyneLogin()
	}
}
