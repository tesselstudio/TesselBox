package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"kaijuengine.com/engine"
)

// ModernFyneUI provides a modern webview-based UI using Fyne
type ModernFyneUI struct {
	app    fyne.App
	window fyne.Window
	host   *engine.Host
}

// NewModernFyneUI creates a new modern Fyne UI
func NewModernFyneUI(host *engine.Host) *ModernFyneUI {
	fyneApp := app.New()

	ui := &ModernFyneUI{
		app:  fyneApp,
		host: host,
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

	// Modern login form
	titleLabel := widget.NewLabel("Welcome to TesselBox")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	subtitleLabel := widget.NewLabel("Sign in to your account")
	subtitleLabel.TextStyle = fyne.TextStyle{}

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username or Email")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	loginBtn := widget.NewButton("Sign In", func() {
		// TODO: Integrate with existing authentication
		println("Login:", usernameEntry.Text)
	})

	githubBtn := widget.NewButton("Continue with GitHub", func() {
		// TODO: Integrate with GitHub OAuth
		println("GitHub OAuth")
	})

	// Modern card layout
	loginCard := container.NewVBox(
		titleLabel,
		subtitleLabel,
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

// Run starts the Fyne application
func (f *ModernFyneUI) Run() {
	if f.app != nil {
		f.app.Run()
	}
}
