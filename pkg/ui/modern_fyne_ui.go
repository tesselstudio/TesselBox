package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tesselstudio/TesselBox/pkg/opengl"
	"github.com/tesselstudio/TesselBox/pkg/player"
	"github.com/tesselstudio/TesselBox/pkg/types"
	"github.com/tesselstudio/TesselBox/pkg/world"
)

// ModernFyneUI provides a modern webview-based UI using Fyne
type ModernFyneUI struct {
	app            fyne.App
	window         fyne.Window
	windowManager  *WindowManager
	onLoginSuccess func()
}

// NewModernFyneUI creates a new modern Fyne UI
func NewModernFyneUI() *ModernFyneUI {
	fyneApp := app.New()

	ui := &ModernFyneUI{
		app: fyneApp,
	}

	// Create window manager
	ui.windowManager = NewWindowManager(ui)

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

	// Welcome message - no authentication required
	titleLabel := widget.NewLabel("Welcome to TesselBox")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	subtitleLabel := widget.NewLabel("Ready to Play!")
	subtitleLabel.TextStyle = fyne.TextStyle{}

	// Simple welcome card
	welcomeCard := container.NewVBox(
		titleLabel,
		subtitleLabel,
		widget.NewSeparator(),
		widget.NewLabel("No authentication required - just click below to start!"),
		widget.NewSeparator(),
	)

	// Add padding and center
	content := container.NewVBox(
		widget.NewLabel(""), // Top spacer
		welcomeCard,
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

	// Game mode cards with buttons
	singleplayerBtn := widget.NewButton("🎮 Singleplayer", func() {
		println("🎮 Starting Singleplayer Game...")
		f.launchSingleplayer()
	})
	singleplayerBtn.Importance = widget.HighImportance

	multiplayerBtn := widget.NewButton("👥 Multiplayer", func() {
		println("👥 Opening Multiplayer...")
		// TODO: Launch multiplayer interface
	})

	settingsBtn := widget.NewButton("⚙️ Settings", func() {
		println("⚙️ Opening Settings...")
		// TODO: Open settings dialog
	})

	singleplayerCard := container.NewVBox(
		singleplayerBtn,
		widget.NewLabel("Start a new singleplayer world"),
	)

	multiplayerCard := container.NewVBox(
		multiplayerBtn,
		widget.NewLabel("Join or host multiplayer games"),
	)

	settingsCard := container.NewVBox(
		settingsBtn,
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

// launchSingleplayer starts a single-player game
func (f *ModernFyneUI) launchSingleplayer() {
	// Hide the current window
	if f.window != nil {
		f.window.Hide()
	}

	// Create a new game window
	gameWindow := f.app.NewWindow("TesselBox - Single Player")
	gameWindow.Resize(fyne.NewSize(1024, 768))
	gameWindow.CenterOnScreen()

	// Create game content
	content := container.NewVBox(
		widget.NewLabel("🎮 Single Player Game"),
		widget.NewSeparator(),
		widget.NewLabel("World Generation: In Progress..."),
		widget.NewProgressBar(),
		widget.NewSeparator(),
		widget.NewLabel("Controls:"),
		widget.NewLabel("WASD - Move | Mouse - Look | Space - Jump"),
		widget.NewLabel("Left Click - Break | Right Click - Place"),
		widget.NewLabel("E - Inventory | ESC - Pause"),
		widget.NewSeparator(),
		widget.NewButton("🚀 Start Game", func() {
			println("🎮 Game starting...")
			f.startActualGame(gameWindow)
		}),
		widget.NewButton("🔙 Back to Menu", func() {
			gameWindow.Close()
			f.ShowGameSelect()
		}),
	)

	gameWindow.SetContent(content)
	gameWindow.Show()
}

// startActualGame initializes and starts the actual game with OpenGL rendering
func (f *ModernFyneUI) startActualGame(currentWindow fyne.Window) {
	// Close the current Fyne window
	currentWindow.Close()

	// Initialize game systems in a separate goroutine to avoid blocking UI
	go func() {
		// Create OpenGL engine
		engine, err := opengl.NewEngine(1024, 768, "TesselBox - Single Player")
		if err != nil {
			println("❌ Failed to create OpenGL engine:", err.Error())
			return
		}
		defer engine.Cleanup()

		// Create world
		gameWorld := world.NewWorld("TesselBox World", 12345) // Fixed seed for reproducibility

		// Create player
		player := player.NewPlayer(gameWorld)

		// Set player at spawn point
		spawn := gameWorld.GetSpawnPoint()
		safeY := gameWorld.GetSafeSpawnHeight(int(spawn.X), int(spawn.Z))
		player.SetPosition(types.NewVec3(float32(spawn.X), float32(safeY), float32(spawn.Z)))

		println("🎮 Game initialized successfully!")
		println("🌍 World: TesselBox World (Seed: 12345)")
		println("👤 Player position: X=", spawn.X, " Y=", safeY, " Z=", spawn.Z)

		// Game loop
		lastTime := time.Now()
		for !engine.ShouldClose() {
			// Calculate delta time
			currentTime := time.Now()
			deltaTime := currentTime.Sub(lastTime).Seconds()
			lastTime = currentTime

			// Update game logic
			player.Update(deltaTime)
			gameWorld.Update(deltaTime, world.NewVec3(0, 0, 0))

			// Update camera from player
			playerPos := player.GetPosition()
			playerRot := player.GetRotation()
			engine.UpdateCameraFromPlayer(
				mgl32.Vec3{playerPos.X, playerPos.Y, playerPos.Z},
				mgl32.Vec3{playerRot.X, playerRot.Y, playerRot.Z},
			)

			// Render
			engine.BeginFrame()
			engine.Render()
			engine.EndFrame()

			// Handle events
			engine.PollEvents()
		}

		println("🎮 Game ended")
	}()
}
