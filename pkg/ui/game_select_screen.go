package ui

import (
	"kaijuengine.com/matrix"
)

// GameSelectScreen represents the game mode selection screen
type GameSelectScreen struct {
	*Menu
	BackgroundPanel    *BackgroundPanel
	TopBar           *Panel
	LogoLabel        *Label
	UserProfile       *UserProfile
	SingleplayerCard  *GameModeCard
	MultiplayerCard   *GameModeCard
	SettingsCard      *GameModeCard
	SignOutButton    *ModernButton
}

// NewGameSelectScreen creates a new game mode selection screen
func NewGameSelectScreen(screenWidth, screenHeight float32) *GameSelectScreen {
	// Create background panel
	background := NewBackgroundPanel("game_select_background", "game_select_bg", screenWidth, screenHeight)
	
	// Create top bar
	topBarHeight := float32(80)
	topBar := NewPanel("top_bar", matrix.NewVec2(0, 0), matrix.NewVec2(screenWidth, topBarHeight))
	topBar.Style.BackgroundColor = matrix.NewColor(20/255.0, 20/255.0, 20/255.0, 0.9)
	topBar.Style.Padding = matrix.NewVec4(20, 20, 20, 20)
	
	// TesselBox logo in top bar
	logoLabel := NewLabel("tesselbox_logo", "TesselBox", matrix.NewVec2(20, 25), matrix.NewVec2(150, 30))
	logoLabel.Style.ForegroundColor = matrix.ColorWhite()
	logoLabel.Style.FontSize = 24
	
	// User profile in top bar
	userProfile := NewUserProfile("user_profile", matrix.NewVec2(screenWidth-250, 20), matrix.NewVec2(230, 40))
	
	// Create game mode cards
	cardWidth := float32(300)
	cardHeight := float32(200)
	cardSpacing := float32(40)
	totalCardsWidth := cardWidth*3 + cardSpacing*2
	startX := (screenWidth - totalCardsWidth) / 2
	cardY := float32(150)
	
	// Singleplayer card
	singleplayerCard := NewGameModeCard(
		"singleplayer_card",
		"singleplayer",
		"singleplayer_icon",
		"Singleplayer",
		"Play offline in your local worlds",
		matrix.NewVec2(startX, cardY),
		matrix.NewVec2(cardWidth, cardHeight),
	)
	
	// Multiplayer card
	multiplayerCard := NewGameModeCard(
		"multiplayer_card",
		"multiplayer",
		"multiplayer_icon",
		"Multiplayer",
		"Join servers and play with friends",
		matrix.NewVec2(startX+cardWidth+cardSpacing, cardY),
		matrix.NewVec2(cardWidth, cardHeight),
	)
	
	// Settings card
	settingsCard := NewGameModeCard(
		"settings_card",
		"settings",
		"settings_icon",
		"Settings",
		"Configure game settings and preferences",
		matrix.NewVec2(startX+(cardWidth+cardSpacing)*2, cardY),
		matrix.NewVec2(cardWidth, cardHeight),
	)
	
	// Sign out button
	signOutButton := NewModernButton(
		"sign_out_button",
		"Sign Out",
		matrix.NewVec2(screenWidth-150, screenHeight-60),
		matrix.NewVec2(120, 40),
	)
	signOutButton.Style.BackgroundColor = matrix.NewColor(239/255.0, 68/255.0, 68/255.0, 1.0) // Red
	
	// Create main panel for menu
	panelWidth := screenWidth
	panelHeight := screenHeight
	menuPanel := NewPanel("game_select_menu_panel", matrix.NewVec2(0, 0), matrix.NewVec2(panelWidth, panelHeight))
	menuPanel.Style.BackgroundColor = matrix.NewColor(0, 0, 0, 0) // Transparent
	
	// Add components to menu panel
	menuPanel.AddChild(topBar.UIComponent)
	menuPanel.AddChild(singleplayerCard.UIComponent)
	menuPanel.AddChild(multiplayerCard.UIComponent)
	menuPanel.AddChild(settingsCard.UIComponent)
	menuPanel.AddChild(signOutButton.UIComponent)
	
	// Add components to top bar
	topBar.AddChild(logoLabel.UIComponent)
	topBar.AddChild(userProfile.UIComponent)
	
	// Create menu structure
	gameSelectScreen := &GameSelectScreen{
		Menu: &Menu{
			ID:       "game_select_screen",
			Type:     MenuTypeGameSelect,
			Title:    "Select Game Mode",
			Panel:    menuPanel,
			Visible:   false,
			Elements:  []UIElement{},
		},
		BackgroundPanel:    background,
		TopBar:           topBar,
		LogoLabel:        logoLabel,
		UserProfile:       userProfile,
		SingleplayerCard:  singleplayerCard,
		MultiplayerCard:   multiplayerCard,
		SettingsCard:      settingsCard,
		SignOutButton:    signOutButton,
	}
	
	// Set up click handlers
	singleplayerCard.OnClick = func() {
		gameSelectScreen.handleSingleplayer()
	}
	
	multiplayerCard.OnClick = func() {
		gameSelectScreen.handleMultiplayer()
	}
	
	settingsCard.OnClick = func() {
		gameSelectScreen.handleSettings()
	}
	
	signOutButton.OnClick = func() {
		gameSelectScreen.handleSignOut()
	}
	
	return gameSelectScreen
}

// handleSingleplayer handles singleplayer selection
func (gss *GameSelectScreen) handleSingleplayer() {
	// This would transition to world selection or direct game start
	// For now, we'll show a simple action
	gss.SingleplayerCard.ModernStyle.Style.BackgroundColor = matrix.NewColor(50/255.0, 50/255.0, 50/255.0, 1.0)
}

// handleMultiplayer handles multiplayer selection
func (gss *GameSelectScreen) handleMultiplayer() {
	// This would transition to server browser or direct connect
	gss.MultiplayerCard.ModernStyle.Style.BackgroundColor = matrix.NewColor(50/255.0, 50/255.0, 50/255.0, 1.0)
}

// handleSettings handles settings selection
func (gss *GameSelectScreen) handleSettings() {
	// This would transition to settings menu
	gss.SettingsCard.ModernStyle.Style.BackgroundColor = matrix.NewColor(50/255.0, 50/255.0, 50/255.0, 1.0)
}

// handleSignOut handles sign out
func (gss *GameSelectScreen) handleSignOut() {
	// This would clear authentication and return to login screen
	// For now, we'll just update user profile
	gss.UserProfile.SetStatus("Disconnected", false)
}

// Show displays game selection screen
func (gss *GameSelectScreen) Show() {
	gss.BackgroundPanel.Panel.SetVisible(true)
	gss.Menu.Panel.SetVisible(true)
	gss.Visible = true
}

// Hide hides game selection screen
func (gss *GameSelectScreen) Hide() {
	gss.BackgroundPanel.Panel.SetVisible(false)
	gss.Menu.Panel.SetVisible(false)
	gss.Visible = false
}

// SetUser sets the current user information
func (gss *GameSelectScreen) SetUser(username string, avatarID string, connected bool) {
	gss.UserProfile.SetUsername(username)
	gss.UserProfile.SetStatus("Connected", connected)
	
	// Update avatar if provided
	if avatarID != "" && gss.UserProfile.Avatar != nil {
		gss.UserProfile.Avatar.SetTextureID(avatarID)
	}
}

// SetHoverState sets hover state for game mode cards
func (gss *GameSelectScreen) SetHoverState(cardType string, isHovered bool) {
	switch cardType {
	case "singleplayer":
		gss.SingleplayerCard.SetHoverState(isHovered)
	case "multiplayer":
		gss.MultiplayerCard.SetHoverState(isHovered)
	case "settings":
		gss.SettingsCard.SetHoverState(isHovered)
	}
}
