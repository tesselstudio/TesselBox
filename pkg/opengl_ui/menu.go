package opengl_ui

import (
	"fmt"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tesselstudio/TesselBox/pkg/game"
	"github.com/tesselstudio/TesselBox/pkg/opengl"
)

// MenuState represents the current menu state
type MenuState int

const (
	MenuStateMain MenuState = iota
	MenuStateSingleplayer
	MenuStateMultiplayer
	MenuStateSettings
)

// OpenGLMenu represents the main menu system
type OpenGLMenu struct {
	window        *glfw.Window
	renderer      *UIRenderer
	engine        *opengl.Engine
	state         MenuState
	shouldRunGame bool
	gameStarted   bool

	// UI components
	rootContainer    *Container
	mainMenu         *Container
	singleplayerMenu *Container
	multiplayerMenu  *Container
	settingsMenu     *Container

	// Game context
	gameController *game.Controller
	gameLoop       *game.GameLoop

	// Input handling
	lastMouseState map[glfw.MouseButton]bool
}

// NewOpenGLMenu creates a new OpenGL-based menu system
func NewOpenGLMenu(engine *opengl.Engine) (*OpenGLMenu, error) {
	window := engine.GetWindow()
	if window == nil {
		return nil, fmt.Errorf("engine window is nil")
	}

	renderer, err := NewUIRenderer(window)
	if err != nil {
		return nil, fmt.Errorf("failed to create UI renderer: %w", err)
	}

	menu := &OpenGLMenu{
		window:         window,
		renderer:       renderer,
		engine:         engine,
		state:          MenuStateMain,
		lastMouseState: make(map[glfw.MouseButton]bool),
	}

	// Set up mouse callback
	fmt.Println("Setting up mouse callback...")
	window.SetMouseButtonCallback(menu.handleMouseClick)
	fmt.Println("Mouse callback set up successfully")

	// Initialize UI components
	if err := menu.initUI(); err != nil {
		return nil, fmt.Errorf("failed to initialize UI: %w", err)
	}

	return menu, nil
}

// initUI initializes all UI components
func (m *OpenGLMenu) initUI() error {
	width, height := m.window.GetSize()

	// Create root container
	m.rootContainer = NewContainer(0, 0, float32(width), float32(height))

	// Create main menu
	m.createMainMenu()

	// Create singleplayer menu
	m.createSingleplayerMenu()

	// Create multiplayer menu
	m.createMultiplayerMenu()

	// Create settings menu
	m.createSettingsMenu()

	// Show main menu by default
	m.rootContainer.AddComponent(m.mainMenu)

	return nil
}

// createMainMenu creates the main menu UI
func (m *OpenGLMenu) createMainMenu() {
	m.mainMenu = NewContainer(0, 0, 800, 600)

	// Center the menu
	windowWidth, windowHeight := m.window.GetSize()
	menuX := float32(windowWidth-800) / 2
	menuY := float32(windowHeight-600) / 2
	m.mainMenu.SetPosition(menuX, menuY)
	fmt.Printf("Menu positioned at (%.1f, %.1f)\n", menuX, menuY)

	// Title
	titleLabel := NewLabel(0, 50, 800, 80, "AIIIIAII")
	titleLabel.SetFontSize(48)
	titleLabel.SetTextColor(mgl32.Vec4{1, 1, 1, 1})
	titleLabel.SetBackground(mgl32.Vec4{0, 0, 0, 0}) // Transparent

	// Singleplayer button (Image)
	singleplayerBtn := NewImageButton(299, 200, 202, 53, "/home/jason/TesselBox/button_singleplayer.png")
	singleplayerBtn.SetOnClick(func() {
		fmt.Println("Singleplayer selected")
		m.state = MenuStateSingleplayer
		m.switchToSingleplayerMenu()
	})

	// Multiplayer button (Image)
	multiplayerBtn := NewImageButton(299, 280, 202, 53, "/home/jason/TesselBox/button_multiplayer.png")
	multiplayerBtn.SetOnClick(func() {
		fmt.Println("Multiplayer selected")
		m.state = MenuStateMultiplayer
		m.switchToMultiplayerMenu()
	})

	// Add components to main menu with vertical layout
	m.mainMenu.AddComponent(titleLabel)
	m.mainMenu.AddComponent(singleplayerBtn)
	m.mainMenu.AddComponent(multiplayerBtn)

	// Set absolute layout to preserve button positions and sizes
	m.mainMenu.SetLayout(NewAbsoluteLayout())
}

// createSingleplayerMenu creates the singleplayer menu UI
func (m *OpenGLMenu) createSingleplayerMenu() {
	m.singleplayerMenu = NewContainer(0, 0, 800, 600)

	// Center the menu
	windowWidth, windowHeight := m.window.GetSize()
	menuX := float32(windowWidth-800) / 2
	menuY := float32(windowHeight-600) / 2
	m.singleplayerMenu.SetPosition(menuX, menuY)

	// Title
	titleLabel := NewLabel(0, 50, 800, 80, "Singleplayer")
	titleLabel.SetFontSize(36)
	titleLabel.SetTextColor(mgl32.Vec4{1, 1, 1, 1})

	// Start new world button (Image)
	startWorldBtn := NewImageButton(299, 200, 202, 53, "/home/jason/TesselBox/button_edit-display-name.png")
	startWorldBtn.SetOnClick(func() {
		fmt.Println("🌍 Starting new world...")
		m.startSingleplayerGame("TesselBox World", 12345)
	})

	// Load world button
	loadWorldBtn := NewButton(200, 280, 400, 60, "📁 Load World")
	loadWorldBtn.SetBackground(mgl32.Vec4{0.6, 0.4, 0.2, 1.0})
	loadWorldBtn.SetOnClick(func() {
		fmt.Println("📁 Load world selected")
		// TODO: Implement world loading
	})

	// Back button
	backBtn := NewButton(200, 440, 400, 60, "🔙 Back")
	backBtn.SetBackground(mgl32.Vec4{0.5, 0.5, 0.5, 1.0})
	backBtn.SetOnClick(func() {
		fmt.Println("🔙 Back to main menu")
		m.state = MenuStateMain
		m.switchToMainMenu()
	})

	// Add components
	m.singleplayerMenu.AddComponent(titleLabel)
	m.singleplayerMenu.AddComponent(startWorldBtn)
	m.singleplayerMenu.AddComponent(loadWorldBtn)
	m.singleplayerMenu.AddComponent(backBtn)

	// Set vertical layout
	m.singleplayerMenu.SetLayout(NewVerticalLayout(20, 0))
}

// createMultiplayerMenu creates the multiplayer menu UI
func (m *OpenGLMenu) createMultiplayerMenu() {
	m.multiplayerMenu = NewContainer(0, 0, 800, 600)

	// Center the menu
	windowWidth, windowHeight := m.window.GetSize()
	menuX := float32(windowWidth-800) / 2
	menuY := float32(windowHeight-600) / 2
	m.multiplayerMenu.SetPosition(menuX, menuY)

	// Title
	titleLabel := NewLabel(0, 50, 800, 80, "Multiplayer")
	titleLabel.SetFontSize(36)
	titleLabel.SetTextColor(mgl32.Vec4{1, 1, 1, 1})

	// Join server button
	joinServerBtn := NewButton(200, 200, 400, 60, "🔗 Join Server")
	joinServerBtn.SetBackground(mgl32.Vec4{0.2, 0.6, 0.2, 1.0})
	joinServerBtn.SetOnClick(func() {
		fmt.Println("🔗 Join server selected")
		// TODO: Implement server joining
	})

	// Host server button
	hostServerBtn := NewButton(200, 280, 400, 60, "🏠 Host Server")
	hostServerBtn.SetBackground(mgl32.Vec4{0.6, 0.4, 0.2, 1.0})
	hostServerBtn.SetOnClick(func() {
		fmt.Println("🏠 Host server selected")
		// TODO: Implement server hosting
	})

	// Back button
	backBtn := NewButton(200, 440, 400, 60, "🔙 Back")
	backBtn.SetBackground(mgl32.Vec4{0.5, 0.5, 0.5, 1.0})
	backBtn.SetOnClick(func() {
		fmt.Println("🔙 Back to main menu")
		m.state = MenuStateMain
		m.switchToMainMenu()
	})

	// Add components
	m.multiplayerMenu.AddComponent(titleLabel)
	m.multiplayerMenu.AddComponent(joinServerBtn)
	m.multiplayerMenu.AddComponent(hostServerBtn)
	m.multiplayerMenu.AddComponent(backBtn)

	// Set vertical layout
	m.multiplayerMenu.SetLayout(NewVerticalLayout(20, 0))
}

// createSettingsMenu creates the settings menu UI
func (m *OpenGLMenu) createSettingsMenu() {
	m.settingsMenu = NewContainer(0, 0, 800, 600)

	// Center the menu
	windowWidth, windowHeight := m.window.GetSize()
	menuX := float32(windowWidth-800) / 2
	menuY := float32(windowHeight-600) / 2
	m.settingsMenu.SetPosition(menuX, menuY)

	// Title
	titleLabel := NewLabel(0, 50, 800, 80, "Settings")
	titleLabel.SetFontSize(36)
	titleLabel.SetTextColor(mgl32.Vec4{1, 1, 1, 1})

	// Video settings button
	videoBtn := NewButton(200, 200, 400, 60, "🎥 Video Settings")
	videoBtn.SetBackground(mgl32.Vec4{0.2, 0.6, 0.2, 1.0})
	videoBtn.SetOnClick(func() {
		fmt.Println("🎥 Video settings selected")
		// TODO: Implement video settings
	})

	// Audio settings button
	audioBtn := NewButton(200, 280, 400, 60, "🔊 Audio Settings")
	audioBtn.SetBackground(mgl32.Vec4{0.6, 0.4, 0.2, 1.0})
	audioBtn.SetOnClick(func() {
		fmt.Println("🔊 Audio settings selected")
		// TODO: Implement audio settings
	})

	// Controls button
	controlsBtn := NewButton(200, 360, 400, 60, "🎮 Controls")
	controlsBtn.SetBackground(mgl32.Vec4{0.4, 0.2, 0.8, 1.0})
	controlsBtn.SetOnClick(func() {
		fmt.Println("🎮 Controls selected")
		// TODO: Implement controls settings
	})

	// Back button
	backBtn := NewButton(200, 440, 400, 60, "🔙 Back")
	backBtn.SetBackground(mgl32.Vec4{0.5, 0.5, 0.5, 1.0})
	backBtn.SetOnClick(func() {
		fmt.Println("🔙 Back to main menu")
		m.state = MenuStateMain
		m.switchToMainMenu()
	})

	// Add components
	m.settingsMenu.AddComponent(titleLabel)
	m.settingsMenu.AddComponent(videoBtn)
	m.settingsMenu.AddComponent(audioBtn)
	m.settingsMenu.AddComponent(controlsBtn)
	m.settingsMenu.AddComponent(backBtn)

	// Set vertical layout
	m.settingsMenu.SetLayout(NewVerticalLayout(20, 0))
}

// switchToMainMenu switches to the main menu
func (m *OpenGLMenu) switchToMainMenu() {
	m.rootContainer.components = m.rootContainer.components[:0] // Clear components
	m.rootContainer.AddComponent(m.mainMenu)
}

// switchToSingleplayerMenu switches to the singleplayer menu
func (m *OpenGLMenu) switchToSingleplayerMenu() {
	m.rootContainer.components = m.rootContainer.components[:0] // Clear components
	m.rootContainer.AddComponent(m.singleplayerMenu)
}

// switchToMultiplayerMenu switches to the multiplayer menu
func (m *OpenGLMenu) switchToMultiplayerMenu() {
	m.rootContainer.components = m.rootContainer.components[:0] // Clear components
	m.rootContainer.AddComponent(m.multiplayerMenu)
}

// switchToSettingsMenu switches to the settings menu
func (m *OpenGLMenu) switchToSettingsMenu() {
	m.rootContainer.components = m.rootContainer.components[:0] // Clear components
	m.rootContainer.AddComponent(m.settingsMenu)
}

// startSingleplayerGame starts a singleplayer game
func (m *OpenGLMenu) startSingleplayerGame(worldName string, seed int64) {
	fmt.Printf("🎮 Starting singleplayer game: %s (seed: %d)\n", worldName, seed)

	// Create game controller
	m.gameController = game.NewController()

	// Create input handler
	inputHandler := game.NewInputHandler(m.window)

	// Create game loop
	m.gameLoop = game.NewGameLoop(m.gameController, m.engine, 60)

	// Lock cursor for game
	inputHandler.LockMouse()

	// Start world
	m.gameController.StartWorld(worldName, seed)

	// Set flags
	m.shouldRunGame = true
	m.gameStarted = true

	// Start game loop in goroutine
	go func() {
		defer m.cleanupGame()
		if err := m.gameLoop.Run(); err != nil {
			fmt.Printf("❌ Game loop error: %v\n", err)
		}
	}()
}

// cleanupGame cleans up game resources
func (m *OpenGLMenu) cleanupGame() {
	if m.gameController != nil {
		m.gameController.Stop()
		m.gameController = nil
	}

	m.gameLoop = nil
	m.gameStarted = false
	m.shouldRunGame = false

	fmt.Println("🎮 Game stopped")
}

// HandleInput handles input events
func (m *OpenGLMenu) HandleInput() {
	if m.gameStarted {
		// Game is running, handle game input
		return
	}

	// Handle menu input
	// This is handled by the UI components in the Update method
}

// handleMouseClick handles mouse button click events
func (m *OpenGLMenu) handleMouseClick(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if m.gameStarted {
		return
	}

	// Get cursor position for UI interaction
	cursorX, cursorY := window.GetCursorPos()
	_, windowHeight := window.GetSize()
	glY := float32(windowHeight) - float32(cursorY)
	glX := float32(cursorX)

	fmt.Printf("Mouse click: button=%v, action=%v, pos=(%.1f,%.1f) -> gl=(%.1f,%.1f)\n",
		button, action, cursorX, cursorY, glX, glY)

	// Handle mouse input for UI components
	handled := m.rootContainer.HandleMouse(glX, glY, button, action)
	fmt.Printf("Mouse event handled: %v\n", handled)
}

// Update handles menu system
func (m *OpenGLMenu) Update() {
	if m.gameStarted {
		// Game is running, don't process menu input
		return
	}

	// Get cursor position for UI interaction
	cursorX, cursorY := m.window.GetCursorPos()

	// Convert to OpenGL coordinates (flip Y)
	_, windowHeight := m.window.GetSize()
	glY := float32(windowHeight) - float32(cursorY)
	glX := float32(cursorX)

	// Handle mouse hover only (clicks are handled by callback)
	m.rootContainer.HandleMouse(glX, glY, 0, glfw.Release)
}

// Render renders the menu system
func (m *OpenGLMenu) Render() {
	if m.gameStarted {
		// Game is running, don't render menu
		return
	}

	// Begin UI rendering
	m.renderer.Begin()

	// Render root container (which contains the current menu)
	m.rootContainer.Render(m.renderer)

	// End UI rendering
	m.renderer.End()
}

// ShouldRunGame returns true if the game should be running
func (m *OpenGLMenu) ShouldRunGame() bool {
	return m.shouldRunGame
}

// IsGameRunning returns true if the game is currently running
func (m *OpenGLMenu) IsGameRunning() bool {
	return m.gameStarted
}

// Cleanup cleans up menu resources
func (m *OpenGLMenu) Cleanup() {
	if m.renderer != nil {
		m.renderer.Cleanup()
	}

	if m.gameController != nil {
		m.gameController.Stop()
	}
}

// Resize handles window resize events
func (m *OpenGLMenu) Resize(width, height int) {
	if m.renderer != nil {
		m.renderer.Resize(width, height)
	}

	// Update container sizes
	m.rootContainer.SetSize(float32(width), float32(height))

	// Recenter menus
	menuX := float32(width-800) / 2
	menuY := float32(height-600) / 2

	m.mainMenu.SetPosition(menuX, menuY)
	m.singleplayerMenu.SetPosition(menuX, menuY)
	m.multiplayerMenu.SetPosition(menuX, menuY)
	m.settingsMenu.SetPosition(menuX, menuY)
}
