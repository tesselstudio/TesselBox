package ui

import (
	"fmt"
)

// ExampleUsage demonstrates how to use the new modern UI system
func ExampleUsage() {
	// Screen dimensions
	screenWidth := float32(1920)
	screenHeight := float32(1080)

	// Create menu manager with modern UI support
	menuManager := NewMenuManager(screenWidth, screenHeight)

	// Show login screen initially
	menuManager.ShowLoginScreen()

	// Simulate successful authentication
	// In a real implementation, this would be called by OAuth callback
	user := &GitHubUser{
		ID:        123,
		Login:     "dev_player",
		Name:      "Developer Player",
		Email:     "dev@example.com",
		AvatarURL: "github_avatar_123",
	}
	menuManager.HandleAuthenticationSuccess(user)

	// Now the game selection screen is visible with user info
	gameSelectScreen := menuManager.GetGameSelectScreen()
	if gameSelectScreen != nil {
		fmt.Println("Game selection screen is now active")
		fmt.Println("User: dev_player")
		fmt.Println("Connected: true")
	}

	// Handle user interactions
	// Singleplayer selection
	menuManager.ShowGameSelectScreen()

	// Sign out
	menuManager.HandleSignOut()

	// Back to login screen
	fmt.Println("User signed out, back to login screen")
}

// CreateModernMenuManager creates a menu manager with modern UI screens
func CreateModernMenuManager(screenWidth, screenHeight float32) *MenuManager {
	menuManager := NewMenuManager(screenWidth, screenHeight)

	// Set up callbacks for modern UI flow
	menuManager.SetOnStartGame(func(worldName string, seed int64) {
		fmt.Printf("Starting game: %s with seed: %d\n", worldName, seed)
	})

	menuManager.SetOnResumeGame(func() {
		fmt.Println("Resuming game")
	})

	menuManager.SetOnQuitToMenu(func() {
		fmt.Println("Quitting to menu")
	})

	menuManager.SetOnQuitToDesktop(func() {
		fmt.Println("Quitting to desktop")
	})

	return menuManager
}

// UIConfiguration represents modern UI configuration
type UIConfiguration struct {
	ScreenWidth  float32
	ScreenHeight float32
	Theme        string // "dark", "light"
	Language     string // "en", "es", etc.
	ShowFPS      bool
	ShowDebug    bool
}

// DefaultUIConfiguration returns default modern UI settings
func DefaultUIConfiguration() *UIConfiguration {
	return &UIConfiguration{
		ScreenWidth:  1920,
		ScreenHeight: 1080,
		Theme:        "dark",
		Language:     "en",
		ShowFPS:      false,
		ShowDebug:    false,
	}
}

// ApplyUITheme applies a theme to UI components
func ApplyUITheme(theme string) {
	switch theme {
	case "dark":
		// Apply dark theme colors
		fmt.Println("Applying dark theme")
	case "light":
		// Apply light theme colors
		fmt.Println("Applying light theme")
	default:
		fmt.Println("Using default theme")
	}
}

// ResponsiveLayout adjusts UI layout based on screen size
func ResponsiveLayout(screenWidth, screenHeight float32) (cardWidth, cardHeight, spacing float32) {
	// Base dimensions for 1920x1080
	baseWidth := float32(300)
	baseHeight := float32(200)
	baseSpacing := float32(40)

	// Scale for different screen sizes
	scaleX := screenWidth / 1920
	scaleY := screenHeight / 1080
	scale := (scaleX + scaleY) / 2

	return baseWidth * scale, baseHeight * scale, baseSpacing * scale
}
