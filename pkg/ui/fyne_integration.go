package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FyneUIWrapper wraps Fyne functionality for integration
type FyneUIWrapper struct {
	app    fyne.App
	window fyne.Window
}

// NewFyneUIWrapper creates a new Fyne UI wrapper
func NewFyneUIWrapper() *FyneUIWrapper {
	fyneApp := app.New()

	wrapper := &FyneUIWrapper{
		app: fyneApp,
	}

	return wrapper
}

// ShowLoginWindow displays a modern login window using Fyne
func (f *FyneUIWrapper) ShowLoginWindow() {
	if f.window != nil {
		f.window.Show()
		return
	}

	f.window = f.app.NewWindow("TesselBox - Login")
	f.window.Resize(fyne.NewSize(400, 500))
	f.window.CenterOnScreen()

	// Create login form
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	loginBtn := widget.NewButton("Login", func() {
		// TODO: Implement login logic
		println("Login clicked:", usernameEntry.Text)
	})

	githubBtn := widget.NewButton("Login with GitHub", func() {
		// TODO: Implement GitHub OAuth
		println("GitHub login clicked")
	})

	// Layout the form
	form := container.NewVBox(
		widget.NewLabel("Welcome to TesselBox"),
		widget.NewSeparator(),
		widget.NewLabel("Username:"),
		usernameEntry,
		widget.NewLabel("Password:"),
		passwordEntry,
		widget.NewSeparator(),
		loginBtn,
		githubBtn,
	)

	f.window.SetContent(form)
	f.window.Show()
}

// Hide closes the Fyne window
func (f *FyneUIWrapper) Hide() {
	if f.window != nil {
		f.window.Hide()
	}
}

// Show displays the Fyne window
func (f *FyneUIWrapper) Show() {
	if f.window != nil {
		f.window.Show()
	}
}

// Run starts the Fyne application
func (f *FyneUIWrapper) Run() {
	if f.app != nil {
		f.app.Run()
	}
}
