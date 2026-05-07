package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"kaijuengine.com/engine"
)

// FyneKaijuBridge integrates Fyne UI with game systems
type FyneKaijuBridge struct {
	app    fyne.App
	window fyne.Window
	host   *engine.Host

	// UI state
	isFyneActive bool
	fyneUI       *ModernFyneUI
}

// NewFyneKaijuBridge creates a new bridge between Fyne and game systems
func NewFyneKaijuBridge(host *engine.Host) *FyneKaijuBridge {
	fyneApp := app.New()

	bridge := &FyneKaijuBridge{
		app:  fyneApp,
		host: host,
	}

	return bridge
}

// ShowFyneLogin displays Fyne login and integrates with Kaiju
func (b *FyneKaijuBridge) ShowFyneLogin() {
	if b.window != nil {
		b.window.Show()
		return
	}

	b.window = b.app.NewWindow("TesselBox - Modern Login")
	b.window.Resize(fyne.NewSize(400, 500))
	b.window.CenterOnScreen()

	// Modern login form
	titleLabel := widget.NewLabel("Welcome to TesselBox")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username or Email")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	loginBtn := widget.NewButton("Sign In", func() {
		b.handleFyneLogin(usernameEntry.Text, passwordEntry.Text)
	})

	githubBtn := widget.NewButton("Continue with GitHub", func() {
		b.handleGitHubLogin()
	})

	// Modern card layout
	loginCard := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		widget.NewLabel("Username"),
		usernameEntry,
		widget.NewLabel("Password"),
		passwordEntry,
		widget.NewSeparator(),
		container.NewHBox(
			loginBtn,
			githubBtn,
		),
	)

	// Add padding and center
	content := container.NewVBox(
		widget.NewLabel(""), // Top spacer
		loginCard,
		widget.NewLabel(""), // Bottom spacer
	)

	b.window.SetContent(content)
	b.window.Show()
	b.isFyneActive = true
}

// ShowFyneGameSelect displays game mode selection
func (b *FyneKaijuBridge) ShowFyneGameSelect() {
	if b.window != nil {
		b.window.Close()
	}

	b.window = b.app.NewWindow("TesselBox - Game Mode")
	b.window.Resize(fyne.NewSize(600, 400))
	b.window.CenterOnScreen()

	titleLabel := widget.NewLabel("Select Game Mode")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Game mode cards with buttons
	singleplayerCard := container.NewVBox(
		widget.NewLabel("🎮 Singleplayer"),
		widget.NewLabel("Start a new singleplayer world"),
		widget.NewButton("Play", func() {
			b.handleGameModeSelect("singleplayer")
		}),
	)

	multiplayerCard := container.NewVBox(
		widget.NewLabel("👥 Multiplayer"),
		widget.NewLabel("Join or host multiplayer games"),
		widget.NewButton("Play", func() {
			b.handleGameModeSelect("multiplayer")
		}),
	)

	settingsCard := container.NewVBox(
		widget.NewLabel("⚙️ Settings"),
		widget.NewLabel("Configure game settings"),
		widget.NewButton("Open", func() {
			b.handleGameModeSelect("settings")
		}),
	)

	gameCards := container.NewHBox(
		singleplayerCard,
		multiplayerCard,
		settingsCard,
	)

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		gameCards,
	)

	b.window.SetContent(content)
	b.window.Show()
}

// handleFyneLogin processes Fyne login and integrates with Kaiju
func (b *FyneKaijuBridge) handleFyneLogin(username, password string) {
	println("Fyne Login:", username)

	// TODO: Integrate with existing Kaiju authentication
	// For now, simulate successful login
	b.onLoginSuccess()
}

// handleGitHubLogin processes GitHub OAuth
func (b *FyneKaijuBridge) handleGitHubLogin() {
	println("GitHub OAuth requested")

	// TODO: Integrate with existing GitHub OAuth system
	// For now, simulate successful login
	b.onLoginSuccess()
}

// handleGameModeSelect processes game mode selection
func (b *FyneKaijuBridge) handleGameModeSelect(mode string) {
	println("Game mode selected:", mode)

	// Hide Fyne window
	b.HideFyne()

	// Handle game modes
	switch mode {
	case "singleplayer":
		println("Starting singleplayer game...")
		// TODO: Start singleplayer game
	case "multiplayer":
		println("Starting multiplayer game...")
		// TODO: Start multiplayer game
	case "settings":
		println("Opening settings...")
		// TODO: Open settings
	}
}

// onLoginSuccess handles successful login
func (b *FyneKaijuBridge) onLoginSuccess() {
	// Hide login window and show game selection
	b.ShowFyneGameSelect()
}

// HideFyne closes Fyne window and switches to Kaiju
func (b *FyneKaijuBridge) HideFyne() {
	if b.window != nil {
		b.window.Hide()
		b.isFyneActive = false
	}
}

// ShowFyne displays Fyne window
func (b *FyneKaijuBridge) ShowFyne() {
	if b.window != nil {
		b.window.Show()
		b.isFyneActive = true
	}
}

// SwitchToKaiju switches back to main menu
func (b *FyneKaijuBridge) SwitchToKaiju() {
	b.HideFyne()
	b.ShowFyneLogin()
}

// IsFyneActive returns true if Fyne UI is currently active
func (b *FyneKaijuBridge) IsFyneActive() bool {
	return b.isFyneActive
}

// Run starts the Fyne application
func (b *FyneKaijuBridge) Run() {
	if b.app != nil {
		b.app.Run()
	}
}
