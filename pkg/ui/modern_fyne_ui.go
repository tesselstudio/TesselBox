package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ModernFyneUI provides a modern webview-based UI using Fyne
type ModernFyneUI struct {
	app            fyne.App
	window         fyne.Window
	onLoginSuccess func()
}

// NewModernFyneUI creates a new modern Fyne UI
func NewModernFyneUI() *ModernFyneUI {
	fyneApp := app.New()

	ui := &ModernFyneUI{
		app: fyneApp,
	}

	return ui
}

// ShowLogin displays a modern login screen
func (f *ModernFyneUI) ShowLogin() {
	if f.window != nil {
		f.window.Show()
		return
	}

	f.window = f.app.NewWindow("TesselBox - Modern Login")
	f.window.Resize(fyne.NewSize(400, 500))
	f.window.CenterOnScreen()

	// Modern login form - GitHub OAuth only
	titleLabel := widget.NewLabel("Welcome to TesselBox")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	subtitleLabel := widget.NewLabel("Sign in with GitHub")
	subtitleLabel.TextStyle = fyne.TextStyle{}

	// GitHub authentication status
	statusLabel := widget.NewLabel("Initializing GitHub OAuth...")
	statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	githubBtn := widget.NewButton("Authenticate with GitHub", func() {
		// Start GitHub OAuth flow
		statusLabel.SetText("Connecting to GitHub...")
		println("🔐 Starting GitHub OAuth authentication")

		// Create GitHub auth handler
		githubAuth := NewGitHubAuth(nil, func(user *GitHubUser, err error) {
			if err != nil {
				statusLabel.SetText("Authentication failed: " + err.Error())
				return
			}

			statusLabel.SetText("✅ Authenticated as " + user.Login)
			println("🎉 GitHub authentication successful!")

			// Close login window and show game selection
			f.Hide()
			if f.onLoginSuccess != nil {
				f.onLoginSuccess()
			}
		})

		// Start GitHub authentication
		githubAuth.Authenticate()
	})

	// Modern card layout - GitHub OAuth only
	loginCard := container.NewVBox(
		titleLabel,
		subtitleLabel,
		widget.NewSeparator(),
		statusLabel,
		widget.NewSeparator(),
		githubBtn,
	)

	// Add padding and center
	content := container.NewVBox(
		widget.NewLabel(""), // Top spacer
		loginCard,
		widget.NewLabel(""), // Bottom spacer
	)

	f.window.SetContent(content)
	f.window.Show()
}

// ShowGameSelect displays game mode selection
func (f *ModernFyneUI) ShowGameSelect() {
	if f.window != nil {
		f.window.Show()
		return
	}

	f.window = f.app.NewWindow("TesselBox - Game Mode")
	f.window.Resize(fyne.NewSize(600, 400))
	f.window.CenterOnScreen()

	titleLabel := widget.NewLabel("Select Game Mode")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Game mode cards
	singleplayerCard := container.NewVBox(
		widget.NewLabel("🎮 Singleplayer"),
		widget.NewLabel("Singleplayer"),
		widget.NewLabel("Start a new singleplayer world"),
	)

	multiplayerCard := container.NewVBox(
		widget.NewLabel("👥 Multiplayer"),
		widget.NewLabel("Multiplayer"),
		widget.NewLabel("Join or host multiplayer games"),
	)

	settingsCard := container.NewVBox(
		widget.NewLabel("⚙️ Settings"),
		widget.NewLabel("Settings"),
		widget.NewLabel("Configure game settings"),
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

	f.window.SetContent(content)
	f.window.Show()
}

// Run starts the Fyne application
func (f *ModernFyneUI) Run() {
	if f.app != nil {
		f.app.Run()
	}
}

// Hide closes the Fyne window
func (f *ModernFyneUI) Hide() {
	if f.window != nil {
		f.window.Hide()
	}
}

// Show displays the Fyne window
func (f *ModernFyneUI) Show() {
	if f.window != nil {
		f.window.Show()
	}
}
