package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"kaijuengine.com/engine"
)

// FyneExample demonstrates how to use the new Fyne UI system
type FyneExample struct {
	app    fyne.App
	window fyne.Window
	host   *engine.Host
}

// NewFyneExample creates a new Fyne example
func NewFyneExample(host *engine.Host) *FyneExample {
	fyneApp := app.New()

	example := &FyneExample{
		app:  fyneApp,
		host: host,
	}

	return example
}

// ShowModernLogin displays a modern login screen using Fyne
func (f *FyneExample) ShowModernLogin() {
	if f.window != nil {
		f.window.Show()
		return
	}

	f.window = f.app.NewWindow("TesselBox - Modern Login")
	f.window.Resize(fyne.NewSize(800, 600))
	f.window.CenterOnScreen()

	// Create modern login form with better styling
	titleLabel := widget.NewLabel("Welcome to TesselBox")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Enter your username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Enter your password")

	loginBtn := widget.NewButton("Login", func() {
		// TODO: Integrate with existing authentication system
		println("Login attempt:", usernameEntry.Text)
	})

	githubBtn := widget.NewButton("Continue with GitHub", func() {
		// TODO: Integrate with GitHub OAuth
		println("GitHub OAuth requested")
	})

	// Create modern card-style layout
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

	// Add some padding and styling
	paddedContent := container.NewVBox(
		widget.NewLabel(""), // Spacer for top padding
		loginCard,
		widget.NewLabel(""), // Spacer for bottom padding
	)

	f.window.SetContent(paddedContent)
	f.window.Show()
}

// Hide closes the Fyne window
func (f *FyneExample) Hide() {
	if f.window != nil {
		f.window.Hide()
	}
}

// Show displays the Fyne window
func (f *FyneExample) Show() {
	if f.window != nil {
		f.window.Show()
	}
}

// Run starts the Fyne application
func (f *FyneExample) Run() {
	if f.app != nil {
		f.app.Run()
	}
}
