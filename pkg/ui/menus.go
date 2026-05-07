package ui

import (
	"sync"
)

// GitHubUser represents a GitHub user profile
type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// SimpleMenuManager manages Fyne UI only
type SimpleMenuManager struct {
	mu          sync.RWMutex
	fyneBridge  *FyneKaijuBridge
	onStartGame func(worldName string, seed int64)
	onQuitGame  func()
}

// NewSimpleMenuManager creates a new Fyne-only menu manager
func NewSimpleMenuManager() *SimpleMenuManager {
	return &SimpleMenuManager{}
}

// SetFyneBridge sets the Fyne bridge
func (sm *SimpleMenuManager) SetFyneBridge(bridge *FyneKaijuBridge) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.fyneBridge = bridge
}

// SetOnStartGame sets the callback for starting a game
func (sm *SimpleMenuManager) SetOnStartGame(callback func(worldName string, seed int64)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onStartGame = callback
}

// SetOnQuitGame sets the callback for quitting the game
func (sm *SimpleMenuManager) SetOnQuitGame(callback func()) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onQuitGame = callback
}

// ShowLogin shows the Fyne login screen
func (sm *SimpleMenuManager) ShowLogin() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if sm.fyneBridge != nil {
		sm.fyneBridge.ShowFyneLogin()
	}
}

// ShowGameSelect shows the Fyne game selection screen
func (sm *SimpleMenuManager) ShowGameSelect() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if sm.fyneBridge != nil {
		sm.fyneBridge.ShowFyneGameSelect()
	}
}

// HandleGameMode handles game mode selection
func (sm *SimpleMenuManager) HandleGameMode(mode string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	switch mode {
	case "singleplayer":
		if sm.onStartGame != nil {
			sm.onStartGame("New World", 12345)
		}
	case "multiplayer":
		if sm.onStartGame != nil {
			sm.onStartGame("Multiplayer World", 54321)
		}
	case "settings":
		// Handle settings
	case "quit":
		if sm.onQuitGame != nil {
			sm.onQuitGame()
		}
	}
}
