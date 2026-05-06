package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"kaijuengine.com/bootstrap"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"

	"github.com/tesselstudio/TesselBox/pkg/audio"
	"github.com/tesselstudio/TesselBox/pkg/content"
	"github.com/tesselstudio/TesselBox/pkg/crafting"
	"github.com/tesselstudio/TesselBox/pkg/game"
	"github.com/tesselstudio/TesselBox/pkg/network"
	"github.com/tesselstudio/TesselBox/pkg/oauth"
	"github.com/tesselstudio/TesselBox/pkg/world"
)

// Version is set at build time
var Version = "dev"

// TesselBoxGame represents the main game state
type TesselBoxGame struct {
	host        *engine.Host
	controller  *game.Controller
	updateId    engine.UpdateId
	currentDoc  interface{}
	stateMutex  sync.RWMutex
	oauthConfig *oauth.Config
}

// PluginRegistry returns the plugin types for this game
func (g *TesselBoxGame) PluginRegistry() []reflect.Type {
	return []reflect.Type{}
}

// ContentDatabase returns the game content database
func (g *TesselBoxGame) ContentDatabase() (assets.Database, error) {
	// Use the same approach as Kaiju Engine's test example
	gameContentPath := "game_content"
	if _, err := os.Stat(gameContentPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// Copy content from Kaiju Engine's embedded content
		if err := g.copyEditorContent(); err != nil {
			return nil, err
		}
	}
	// Use GameContentDatabase which handles path transformations for shaders, materials, etc.
	return content.NewGameContentDatabase(gameContentPath)
}

// copyEditorContent copies content from Kaiju Engine's embedded content
func (g *TesselBoxGame) copyEditorContent() error {
	editorContentPath := "/home/jason/TesselBox/kaiju/src/editor/editor_embedded_content/editor_content"
	gameContentPath := "game_content"

	// Create game_content directory
	if err := os.MkdirAll(gameContentPath, 0755); err != nil {
		return err
	}

	// Copy all content from editor_content to game_content
	return filepath.Walk(editorContentPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(editorContentPath, path)
		if err != nil {
			return err
		}

		outPath := filepath.Join(gameContentPath, relPath)

		if info.IsDir() {
			return os.MkdirAll(outPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(outPath, data, info.Mode())
	})
}

// EmbeddedContentDatabase wraps the Kaiju Engine's embedded content
type EmbeddedContentDatabase struct {
	cache map[string][]byte
}

func NewEmbeddedContentDatabase() *EmbeddedContentDatabase {
	return &EmbeddedContentDatabase{
		cache: make(map[string][]byte),
	}
}

func (db *EmbeddedContentDatabase) PostWindowCreate(handle assets.PostWindowCreateHandle) error {
	// No-op for embedded content
	return nil
}

func (db *EmbeddedContentDatabase) Cache(key string, data []byte) {
	db.cache[key] = data
}

func (db *EmbeddedContentDatabase) CacheRemove(key string) {
	delete(db.cache, key)
}

func (db *EmbeddedContentDatabase) CacheClear() {
	db.cache = make(map[string][]byte)
}

func (db *EmbeddedContentDatabase) Read(key string) ([]byte, error) {
	// Check cache first
	if data, exists := db.cache[key]; exists {
		return data, nil
	}

	// Try to read from embedded file system first
	data, err := project_file_system.EngineFS.ReadFile(key)
	if err == nil {
		db.Cache(key, data) // Cache for future reads
		return data, nil
	}

	// Fallback to file system
	data, err = os.ReadFile(key)
	if err == nil {
		db.Cache(key, data) // Cache for future reads
	}
	return data, err
}

func (db *EmbeddedContentDatabase) ReadText(key string) (string, error) {
	data, err := db.Read(key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (db *EmbeddedContentDatabase) Exists(key string) bool {
	// Check cache first
	if _, exists := db.cache[key]; exists {
		return true
	}

	// Check embedded file system first
	if _, err := project_file_system.EngineFS.Open(key); err == nil {
		return true
	}

	// Fallback to file system
	_, err := os.Stat(key)
	return err == nil
}

func (db *EmbeddedContentDatabase) Close() {
	// Clean up cache
	db.CacheClear()
}

// uiManager is the UI manager for the game
var uiManager ui.Manager

// Launch initializes the game
func (g *TesselBoxGame) Launch(host *engine.Host) {
	println("DEBUG: Launch starting...")
	g.host = host

	// Initialize OAuth configuration
	oauthConfig, err := oauth.LoadConfig()
	if err != nil {
		println("DEBUG: OAuth config not found, OAuth login will be disabled:", err.Error())
		g.oauthConfig = nil
	} else {
		g.oauthConfig = oauthConfig
		println("DEBUG: OAuth configuration loaded")
	}

	// Initialize UI Manager
	println("DEBUG: Initializing UI Manager...")
	uiManager.Init(host)
	println("DEBUG: UI Manager initialized")

	// Initialize audio system
	println("DEBUG: Initializing audio...")
	audioManager := audio.GetManager()
	if err := audioManager.Initialize(); err != nil {
		// Audio is optional, log but continue
		println("Warning: Audio initialization failed:", err.Error())
	}

	// Create game controller
	println("DEBUG: Creating controller...")
	g.controller = game.NewController(host)
	println("DEBUG: Controller created")

	// Register state change callback
	g.registerStateCallbacks()

	// Show login screen (create UI before showing window)
	println("DEBUG: Showing login screen...")
	g.showLoginScreen()
	println("DEBUG: Login screen shown")

	// Show the window after UI is ready - critical for window to appear with content
	println("DEBUG: About to show window...")
	if host.Window != nil {
		host.Window.Show()
		println("DEBUG: Window shown")
	} else {
		println("DEBUG: Window is nil!")
	}

	// Register update function for game loop
	println("DEBUG: Registering update...")
	g.updateId = host.Updater.AddUpdate(g.update)
	println("DEBUG: Launch complete")
}

// playUIClick plays the UI click sound
func (g *TesselBoxGame) playUIClick() {
	audio.GetManager().TriggerGameEvent(audio.EventUIClick)
}

// showLoginScreen displays the login UI
func (g *TesselBoxGame) showLoginScreen() {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	uiManager.Clear()

	loginHTML, err := g.host.AssetDatabase().ReadText("ui/login.html")
	if err != nil {
		g.createFallbackLoginScreen()
		return
	}

	doc := markup.DocumentFromHTMLString(&uiManager, string(loginHTML), "", nil, nil, nil)
	g.currentDoc = doc

	// Activate top-level elements first, then all elements
	// This ensures proper hierarchy activation for rendering
	for i := range doc.TopElements {
		if doc.TopElements[i].UI != nil && doc.TopElements[i].UI.Entity() != nil {
			doc.TopElements[i].UI.Entity().Activate()
		}
	}
	for i := range doc.Elements {
		if doc.Elements[i].UI != nil && doc.Elements[i].UI.Entity() != nil {
			doc.Elements[i].UI.Entity().Activate()
		}
	}

	// Setup login handlers
	loginButton, ok := doc.GetElementById("loginButton")
	if ok {
		btn := loginButton.UI.ToButton()
		btn.Base().AddEvent(ui.EventTypeClick, func() {
			g.playUIClick()
			g.transitionToMainMenu()
		})
	}

	// Setup GitHub OAuth login button
	githubLoginButton, ok := doc.GetElementById("githubLoginButton")
	if ok {
		btn := githubLoginButton.UI.ToButton()
		btn.Base().AddEvent(ui.EventTypeClick, func() {
			g.playUIClick()
			g.initiateGitHubOAuth()
		})
	}

	// Focus username input
	if usernameInput, ok := doc.GetElementById("usernameInput"); ok {
		userInput := usernameInput.UI.ToInput()
		userInput.Focus()
	}

	// Add ESC handler for quit - defer to ensure window is ready
	g.host.RunAfterFrames(1, func() {
		if g.host != nil && g.host.Window != nil {
			g.host.Window.Keyboard.AddKeyCallback(func(keyId int, keyState hid.KeyState) {
				if keyId == hid.KeyboardKeyEscape && keyState == hid.KeyStateDown {
					g.host.Close()
				}
			})
		}
	})
}

// transitionToMainMenu transitions from login to main menu
func (g *TesselBoxGame) transitionToMainMenu() {
	if g.controller.TransitionTo(game.GameStateMainMenu) {
		g.showMainMenu()
	}
}

// showMainMenu displays the main menu
func (g *TesselBoxGame) showMainMenu() {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	uiManager.Clear()

	menuHTML, err := g.host.AssetDatabase().ReadText("ui/main_menu.html")
	if err != nil {
		// Fallback
		g.createFallbackMainMenu()
		return
	}

	doc := markup.DocumentFromHTMLString(&uiManager, string(menuHTML), "", nil, nil, nil)
	g.currentDoc = doc

	// Activate top-level elements first, then all elements
	for i := range doc.TopElements {
		if doc.TopElements[i].UI != nil && doc.TopElements[i].UI.Entity() != nil {
			doc.TopElements[i].UI.Entity().Activate()
		}
	}
	for i := range doc.Elements {
		if doc.Elements[i].UI != nil && doc.Elements[i].UI.Entity() != nil {
			doc.Elements[i].UI.Entity().Activate()
		}
	}

	// Single Player button
	if btn, ok := doc.GetElementById("btnSinglePlayer"); ok {
		btn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.playUIClick()
			g.showWorldSelect()
		})
	}

	// Multiplayer button
	if btn, ok := doc.GetElementById("btnMultiplayer"); ok {
		btn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.showMultiplayerMenu()
		})
	}

	// Settings button
	if btn, ok := doc.GetElementById("btnSettings"); ok {
		btn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.showSettings()
		})
	}

	// Quit button
	if btn, ok := doc.GetElementById("btnQuit"); ok {
		btn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.playUIClick()
			g.host.Close()
		})
	}
}

// createFallbackMainMenu creates a basic menu programmatically
func (g *TesselBoxGame) createFallbackMainMenu() {
	panel := uiManager.Add().ToPanel()
	panel.Init(nil, ui.ElementTypePanel)
	panel.SetColor(matrix.Color{0.1, 0.12, 0.18, 1.0})
	panel.Base().Entity().Activate()

	// Title
	titleLabel := uiManager.Add().ToLabel()
	titleLabel.Init("TesselBox")
	titleLabel.SetColor(matrix.ColorWhite())
	titleLabel.SetFontSize(32)
	titleLabel.Base().Entity().Activate()
	panel.AddChild(titleLabel.Base())

	// Single Player button
	btnSP := uiManager.Add().ToButton()
	btnSP.Init(nil, "Single Player")
	btnSP.SetColor(matrix.Color{0.91, 0.27, 0.38, 1.0})
	btnSP.Base().Entity().Activate()
	btnSP.Base().AddEvent(ui.EventTypeClick, func() {
		g.showWorldSelect()
	})
	panel.AddChild(btnSP.Base())

	// Quit button
	btnQuit := uiManager.Add().ToButton()
	btnQuit.Init(nil, "Quit")
	btnQuit.SetColor(matrix.Color{0.5, 0.5, 0.5, 1.0})
	btnQuit.Base().Entity().Activate()
	btnQuit.Base().AddEvent(ui.EventTypeClick, func() {
		g.host.Close()
	})
	panel.AddChild(btnQuit.Base())
}

// showWorldSelect displays the world selection screen
func (g *TesselBoxGame) showWorldSelect() {
	if !g.controller.TransitionTo(game.GameStateWorldSelect) {
		return
	}

	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	uiManager.Clear()

	worldHTML, err := g.host.AssetDatabase().ReadText("ui/world_select.html")
	if err != nil {
		// Fallback: directly start a new world
		g.startNewWorld("New World", time.Now().Unix())
		return
	}

	doc := markup.DocumentFromHTMLString(&uiManager, string(worldHTML), "", nil, nil, nil)
	g.currentDoc = doc

	// Activate top-level elements first, then all elements
	for i := range doc.TopElements {
		if doc.TopElements[i].UI != nil && doc.TopElements[i].UI.Entity() != nil {
			doc.TopElements[i].UI.Entity().Activate()
		}
	}
	for i := range doc.Elements {
		if doc.Elements[i].UI != nil && doc.Elements[i].UI.Entity() != nil {
			doc.Elements[i].UI.Entity().Activate()
		}
	}

	// Load saved worlds
	worlds, _ := world.ListWorlds()

	// World list container and empty state
	worldListEl, hasList := doc.GetElementById("worldList")
	emptyStateEl, hasEmpty := doc.GetElementById("emptyState")
	playBtn, hasPlayBtn := doc.GetElementById("playBtn")

	// Track selected world
	var selectedWorld string

	// Populate world list if we have worlds
	if len(worlds) > 0 && hasList {
		// Hide empty state
		if hasEmpty {
			emptyStateEl.UI.Entity().Deactivate()
		}

		// Enable play button by activating it
		if hasPlayBtn {
			playBtn.UI.Entity().Activate()
		}

		// Create world items
		for i, worldName := range worlds {
			g.createWorldListItem(worldListEl.UI, worldName, i, &selectedWorld)
		}
	}

	// Back button
	if btn, ok := doc.GetElementById("backButton"); ok {
		btn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.playUIClick()
			g.controller.GoToMainMenu()
			g.showMainMenu()
		})
	}

	// Create World button - opens modal
	if btn, ok := doc.GetElementById("createWorldBtn"); ok {
		btn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.playUIClick()
			if modal, ok := doc.GetElementById("createWorldModal"); ok {
				modal.UI.Entity().Activate()
			}
		})
	}

	// Cancel create button
	if btn, ok := doc.GetElementById("cancelCreateBtn"); ok {
		btn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.playUIClick()
			if modal, ok := doc.GetElementById("createWorldModal"); ok {
				modal.UI.Entity().Deactivate()
			}
		})
	}

	// Random seed button
	if btn, ok := doc.GetElementById("randomSeedBtn"); ok {
		btn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.playUIClick()
			if seedInput, ok := doc.GetElementById("worldSeedInput"); ok {
				seedInput.UI.ToInput().SetText(fmt.Sprintf("%d", time.Now().Unix()))
			}
		})
	}

	// Confirm create button
	if btn, ok := doc.GetElementById("confirmCreateBtn"); ok {
		btn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.playUIClick()

			var worldName string
			var seed int64 = time.Now().Unix()

			if nameInput, ok := doc.GetElementById("worldNameInput"); ok {
				worldName = nameInput.UI.ToInput().Text()
			}

			if seedInput, ok := doc.GetElementById("worldSeedInput"); ok {
				seedStr := seedInput.UI.ToInput().Text()
				if seedStr != "" {
					fmt.Sscanf(seedStr, "%d", &seed)
				}
			}

			if worldName == "" {
				worldName = "New World"
			}

			// Ensure unique name
			baseName := worldName
			suffix := 1
			for world.WorldExists(worldName) {
				worldName = fmt.Sprintf("%s (%d)", baseName, suffix)
				suffix++
			}

			g.startNewWorld(worldName, seed)
		})
	}

	// Play button - loads selected world
	if btn, ok := doc.GetElementById("playBtn"); ok {
		btn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.playUIClick()
			if selectedWorld != "" {
				g.loadExistingWorld(selectedWorld)
			} else if len(worlds) > 0 {
				// Load first world if none selected
				g.loadExistingWorld(worlds[0])
			}
		})
	}
}

// createWorldListItem creates a world item in the list
func (g *TesselBoxGame) createWorldListItem(listEl *ui.UI, worldName string, index int, selectedWorld *string) {
	// Create container for world item
	itemPanel := uiManager.Add().ToPanel()
	itemPanel.Init(nil, ui.ElementTypePanel)
	itemPanel.SetColor(matrix.Color{0.15, 0.18, 0.25, 1.0})
	itemPanel.Base().Layout().SetMargin(0, 0, 12, 0)
	itemPanel.Base().Entity().Activate()

	// World name label
	nameLabel := uiManager.Add().ToLabel()
	nameLabel.Init(worldName)
	nameLabel.SetColor(matrix.ColorWhite())
	nameLabel.SetFontSize(18)
	nameLabel.Base().Entity().Activate()
	itemPanel.AddChild(nameLabel.Base())

	// Try to load world info
	var infoText string
	saveManager, err := world.NewSaveManager(worldName)
	if err == nil {
		if info, err := saveManager.LoadWorldInfo(); err == nil {
			lastPlayed := time.Unix(info.LastPlayed, 0).Format("Jan 2, 2006")
			infoText = fmt.Sprintf("Seed: %d | Last played: %s", info.Seed, lastPlayed)
		}
		saveManager.Close()
	}

	if infoText == "" {
		infoText = "No save data"
	}

	// Info label
	infoLabel := uiManager.Add().ToLabel()
	infoLabel.Init(infoText)
	infoLabel.SetColor(matrix.Color{0.6, 0.6, 0.6, 1.0})
	infoLabel.SetFontSize(12)
	infoLabel.Base().Entity().Activate()
	itemPanel.AddChild(infoLabel.Base())

	// Add click handler for selection
	itemPanel.Base().AddEvent(ui.EventTypeClick, func() {
		g.playUIClick()
		*selectedWorld = worldName
		// Visual feedback would go here in a full implementation
	})

	// Add to list
	listEl.ToPanel().AddChild(itemPanel.Base())
}

// loadExistingWorld loads an existing saved world
func (g *TesselBoxGame) loadExistingWorld(worldName string) {
	// Get world info to retrieve seed
	saveManager, err := world.NewSaveManager(worldName)
	if err != nil {
		// Fallback to new world
		g.startNewWorld(worldName, time.Now().Unix())
		return
	}

	var seed int64 = time.Now().Unix()
	if info, err := saveManager.LoadWorldInfo(); err == nil {
		seed = info.Seed
	}
	saveManager.Close()

	// Start world - controller will detect it exists and load data
	g.controller.TransitionTo(game.GameStateLoading)
	g.showLoadingScreen("Loading World...")

	go func() {
		g.controller.StartWorld(worldName, seed)

		g.host.RunAfterTime(time.Millisecond*100, func() {
			g.showGame()
		})
	}()
}

// showMultiplayerMenu displays the multiplayer menu
func (g *TesselBoxGame) showMultiplayerMenu() {
	if !g.controller.TransitionTo(game.GameStateMultiplayer) {
		return
	}

	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	uiManager.Clear()

	// Load multiplayer HTML UI
	mpHTML, err := g.host.AssetDatabase().ReadText("ui/multiplayer.html")
	if err != nil {
		// Fallback to programmatic UI
		g.showFallbackMultiplayerMenu()
		return
	}

	doc := markup.DocumentFromHTMLString(&uiManager, string(mpHTML), "", nil, nil, nil)
	// Activate top-level elements first, then all elements
	for i := range doc.TopElements {
		if doc.TopElements[i].UI != nil && doc.TopElements[i].UI.Entity() != nil {
			doc.TopElements[i].UI.Entity().Activate()
		}
	}
	for i := range doc.Elements {
		if doc.Elements[i].UI != nil && doc.Elements[i].UI.Entity() != nil {
			doc.Elements[i].UI.Entity().Activate()
		}
	}

	// Setup back button
	if backBtn, ok := doc.GetElementById("backBtn"); ok {
		backBtn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.controller.GoToMainMenu()
			g.showMainMenu()
		})
	}

	// Setup connect button
	if connectBtn, ok := doc.GetElementById("connectBtn"); ok {
		connectBtn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.connectToServer()
		})
	}

	// Setup host button
	if hostBtn, ok := doc.GetElementById("hostBtn"); ok {
		hostBtn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.hostServer()
		})
	}

	g.currentDoc = doc
}

// showFallbackMultiplayerMenu shows simple multiplayer UI when HTML fails
func (g *TesselBoxGame) showFallbackMultiplayerMenu() {
	panel := uiManager.Add().ToPanel()
	panel.Init(nil, ui.ElementTypePanel)
	panel.SetColor(matrix.Color{0.1, 0.12, 0.18, 1.0})
	panel.Base().Entity().Activate()

	title := uiManager.Add().ToLabel()
	title.Init("Multiplayer")
	title.SetColor(matrix.ColorWhite())
	title.SetFontSize(28)
	title.Base().Entity().Activate()
	panel.AddChild(title.Base())

	// Server address input
	addrLabel := uiManager.Add().ToLabel()
	addrLabel.Init("Server: localhost:8080")
	addrLabel.SetColor(matrix.Color{0.7, 0.7, 0.7, 1.0})
	addrLabel.SetFontSize(14)
	addrLabel.Base().Entity().Activate()
	panel.AddChild(addrLabel.Base())

	btnConnect := uiManager.Add().ToButton()
	btnConnect.Init(nil, "Connect")
	btnConnect.SetColor(matrix.Color{0.3, 0.6, 1.0, 1.0})
	btnConnect.Base().Entity().Activate()
	btnConnect.Base().AddEvent(ui.EventTypeClick, func() {
		g.connectToServer()
	})
	panel.AddChild(btnConnect.Base())

	btnHost := uiManager.Add().ToButton()
	btnHost.Init(nil, "Host Server")
	btnHost.SetColor(matrix.Color{0.2, 0.8, 0.4, 1.0})
	btnHost.Base().Entity().Activate()
	btnHost.Base().AddEvent(ui.EventTypeClick, func() {
		g.hostServer()
	})
	panel.AddChild(btnHost.Base())

	btnBack := uiManager.Add().ToButton()
	btnBack.Init(nil, "Back")
	btnBack.SetColor(matrix.Color{0.5, 0.5, 0.5, 1.0})
	btnBack.Base().Entity().Activate()
	btnBack.Base().AddEvent(ui.EventTypeClick, func() {
		g.controller.GoToMainMenu()
		g.showMainMenu()
	})
	panel.AddChild(btnBack.Base())
}

// connectToServer connects to a multiplayer server
func (g *TesselBoxGame) connectToServer() {
	// Get server address from UI or use default
	address := "localhost:8080"

	// Start loading
	g.controller.StartLoading()
	g.showLoadingScreen("Connecting to server...")

	// Connect in background
	go func() {
		config := network.ClientConfig{
			ServerAddr: address,
			PlayerName: "Player",
			TickRate:   20,
		}
		client := network.NewClient(config)
		if err := client.Connect(address); err != nil {
			// Connection failed - return to multiplayer menu
			g.host.RunAfterTime(time.Millisecond*100, func() {
				g.showMultiplayerMenu()
			})
			return
		}

		// Connection successful - start game
		g.host.RunAfterTime(time.Millisecond*100, func() {
			g.controller.SetNetworkClient(client)
			g.showGame()
		})
	}()
}

// hostServer starts a multiplayer server
func (g *TesselBoxGame) hostServer() {
	// Start loading
	g.controller.StartLoading()
	g.showLoadingScreen("Starting server...")

	// Host in background
	go func() {
		config := network.ServerConfig{
			Port:         8080,
			MaxPlayers:   8,
			TickRate:     20,
			WorldSize:    16,
			ChunkSize:    16,
			ViewDistance: 8,
		}
		server := network.NewServer(config)
		if err := server.Start(); err != nil {
			// Server failed to start
			g.host.RunAfterTime(time.Millisecond*100, func() {
				g.showMultiplayerMenu()
			})
			return
		}

		// Server started - connect as host and start game
		g.host.RunAfterTime(time.Millisecond*100, func() {
			g.controller.SetNetworkServer(server)
			// Also connect as client
			clientConfig := network.ClientConfig{
				ServerAddr: "localhost:8080",
				PlayerName: "Host",
				TickRate:   20,
			}
			client := network.NewClient(clientConfig)
			client.Connect("localhost:8080")
			g.controller.SetNetworkClient(client)
			g.showGame()
		})
	}()
}

// showLoadingScreen shows a loading screen with a message
func (g *TesselBoxGame) showLoadingScreen(message string) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	uiManager.Clear()

	loadingPanel := uiManager.Add().ToPanel()
	loadingPanel.Init(nil, ui.ElementTypePanel)
	loadingPanel.SetColor(matrix.Color{0.0, 0.0, 0.0, 1.0})
	loadingPanel.Base().Entity().Activate()

	loadingLabel := uiManager.Add().ToLabel()
	loadingLabel.Init(message)
	loadingLabel.SetColor(matrix.ColorWhite())
	loadingLabel.SetFontSize(24)
	loadingLabel.Base().Entity().Activate()
	loadingPanel.AddChild(loadingLabel.Base())
}

// showSettings displays the settings screen
func (g *TesselBoxGame) showSettings() {
	if !g.controller.TransitionTo(game.GameStateSettings) {
		return
	}

	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	uiManager.Clear()

	// Create simple settings menu
	panel := uiManager.Add().ToPanel()
	panel.Init(nil, ui.ElementTypePanel)
	panel.SetColor(matrix.Color{0.1, 0.12, 0.18, 1.0})
	panel.Base().Entity().Activate()

	title := uiManager.Add().ToLabel()
	title.Init("Settings")
	title.SetColor(matrix.ColorWhite())
	title.SetFontSize(28)
	title.Base().Entity().Activate()
	panel.AddChild(title.Base())

	btnBack := uiManager.Add().ToButton()
	btnBack.Init(nil, "Back")
	btnBack.SetColor(matrix.Color{0.5, 0.5, 0.5, 1.0})
	btnBack.Base().Entity().Activate()
	btnBack.Base().AddEvent(ui.EventTypeClick, func() {
		g.controller.CloseSettings()
		if g.controller.IsInMenu() {
			g.showMainMenu()
		} else {
			g.showPauseMenu()
		}
	})
	panel.AddChild(btnBack.Base())
}

// startNewWorld starts a new game world
func (g *TesselBoxGame) startNewWorld(worldName string, seed int64) {
	if !g.controller.TransitionTo(game.GameStateLoading) {
		return
	}

	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	// Show loading screen
	uiManager.Clear()

	loadingPanel := uiManager.Add().ToPanel()
	loadingPanel.Init(nil, ui.ElementTypePanel)
	loadingPanel.SetColor(matrix.Color{0.0, 0.0, 0.0, 1.0})
	loadingPanel.Base().Entity().Activate()

	loadingLabel := uiManager.Add().ToLabel()
	loadingLabel.Init("Loading World...")
	loadingLabel.SetColor(matrix.ColorWhite())
	loadingLabel.SetFontSize(24)
	loadingLabel.Base().Entity().Activate()
	loadingPanel.AddChild(loadingLabel.Base())

	// Start world in background
	go func() {
		g.controller.StartWorld(worldName, seed)

		// Transition to playing state on main thread
		g.host.RunAfterTime(time.Millisecond*100, func() {
			g.showGame()
		})
	}()
}

// showGame shows the in-game screen
func (g *TesselBoxGame) showGame() {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	uiManager.Clear()

	// Load HUD
	hudHTML, err := g.host.AssetDatabase().ReadText("ui/hud.html")
	if err == nil {
		doc := markup.DocumentFromHTMLString(&uiManager, string(hudHTML), "", nil, nil, nil)
		g.currentDoc = doc
		// Activate top-level elements first, then all elements
		for i := range doc.TopElements {
			if doc.TopElements[i].UI != nil && doc.TopElements[i].UI.Entity() != nil {
				doc.TopElements[i].UI.Entity().Activate()
			}
		}
		for i := range doc.Elements {
			if doc.Elements[i].UI != nil && doc.Elements[i].UI.Entity() != nil {
				doc.Elements[i].UI.Entity().Activate()
			}
		}
	}

	// Setup game input
	g.setupGameInput()
}

// showPauseMenu shows the pause menu
func (g *TesselBoxGame) showPauseMenu() {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	pauseHTML, err := g.host.AssetDatabase().ReadText("ui/pause_menu.html")
	if err != nil {
		// Fallback pause menu
		return
	}

	doc := markup.DocumentFromHTMLString(&uiManager, string(pauseHTML), "", nil, nil, nil)
	// Activate top-level elements first, then all elements
	for i := range doc.TopElements {
		if doc.TopElements[i].UI != nil && doc.TopElements[i].UI.Entity() != nil {
			doc.TopElements[i].UI.Entity().Activate()
		}
	}
	for i := range doc.Elements {
		if doc.Elements[i].UI != nil && doc.Elements[i].UI.Entity() != nil {
			doc.Elements[i].UI.Entity().Activate()
		}
	}

	// Resume button
	if btn, ok := doc.GetElementById("btnResume"); ok {
		btn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.controller.TransitionTo(game.GameStatePlaying)
			g.showGame()
		})
	}

	// Settings button
	if btn, ok := doc.GetElementById("btnSettings"); ok {
		btn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.showSettings()
		})
	}

	// Save & Quit to Menu button
	if btn, ok := doc.GetElementById("btnQuitToMenu"); ok {
		btn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.controller.Stop()
			g.controller.GoToMainMenu()
			g.showMainMenu()
		})
	}

	// Save & Quit to Desktop button
	if btn, ok := doc.GetElementById("btnQuitDesktop"); ok {
		btn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.controller.Stop()
			g.host.Close()
		})
	}
}

// setupGameInput configures input handling during gameplay
func (g *TesselBoxGame) setupGameInput() {
	g.host.Window.Keyboard.AddKeyCallback(func(keyId int, keyState hid.KeyState) {
		if keyState != hid.KeyStateDown {
			return
		}

		switch keyId {
		case hid.KeyboardKeyEscape:
			if g.controller.GetState() == game.GameStatePlaying {
				g.controller.TransitionTo(game.GameStatePaused)
				g.showPauseMenu()
			} else if g.controller.GetState() == game.GameStatePaused {
				g.controller.TransitionTo(game.GameStatePlaying)
				g.showGame()
			}
		case hid.KeyboardKeyE:
			if g.controller.GetState() == game.GameStatePlaying {
				g.controller.TransitionTo(game.GameStateInventory)
				g.showInventory()
			} else if g.controller.GetState() == game.GameStateInventory || g.controller.GetState() == game.GameStateCrafting {
				g.controller.TransitionTo(game.GameStatePlaying)
				g.showGame()
			}
		case hid.KeyboardKeyC:
			// Open crafting screen
			if g.controller.GetState() == game.GameStatePlaying {
				g.controller.TransitionTo(game.GameStateCrafting)
				g.showCrafting()
			} else if g.controller.GetState() == game.GameStateCrafting {
				g.controller.TransitionTo(game.GameStatePlaying)
				g.showGame()
			}
		case hid.KeyboardKeyQ:
			// Drop item from selected hotbar slot
			if player := g.controller.GetPlayer(); player != nil {
				player.DropItem()
			}
		case hid.KeyboardKeyF3:
			if g.controller.GetHUD() != nil {
				g.controller.GetHUD().ToggleDebug()
			}
		}

		// Hotbar keys 1-9
		if keyId >= hid.KeyboardKey1 && keyId <= hid.KeyboardKey9 {
			slot := keyId - hid.KeyboardKey1
			if player := g.controller.GetPlayer(); player != nil {
				player.SetHotbarSlot(slot)
			}
		}
	})
}

// showInventory displays the inventory screen
func (g *TesselBoxGame) showInventory() {
	invHTML, err := g.host.AssetDatabase().ReadText("ui/inventory.html")
	if err != nil {
		return
	}

	doc := markup.DocumentFromHTMLString(&uiManager, string(invHTML), "", nil, nil, nil)
	// Activate top-level elements first, then all elements
	for i := range doc.TopElements {
		if doc.TopElements[i].UI != nil && doc.TopElements[i].UI.Entity() != nil {
			doc.TopElements[i].UI.Entity().Activate()
		}
	}
	for i := range doc.Elements {
		if doc.Elements[i].UI != nil && doc.Elements[i].UI.Entity() != nil {
			doc.Elements[i].UI.Entity().Activate()
		}
	}

	// Setup crafting button if it exists
	if craftBtn, ok := doc.GetElementById("craftingBtn"); ok {
		craftBtn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.controller.TransitionTo(game.GameStateCrafting)
			g.showCrafting()
		})
	}

	g.currentDoc = doc
}

// showCrafting displays the crafting screen
func (g *TesselBoxGame) showCrafting() {
	craftHTML, err := g.host.AssetDatabase().ReadText("ui/crafting.html")
	if err != nil {
		// Fallback: show simple crafting UI
		g.showFallbackCrafting()
		return
	}

	doc := markup.DocumentFromHTMLString(&uiManager, string(craftHTML), "", nil, nil, nil)
	// Activate top-level elements first, then all elements
	for i := range doc.TopElements {
		if doc.TopElements[i].UI != nil && doc.TopElements[i].UI.Entity() != nil {
			doc.TopElements[i].UI.Entity().Activate()
		}
	}
	for i := range doc.Elements {
		if doc.Elements[i].UI != nil && doc.Elements[i].UI.Entity() != nil {
			doc.Elements[i].UI.Entity().Activate()
		}
	}

	// Setup close button
	if closeBtn, ok := doc.GetElementById("closeBtn"); ok {
		closeBtn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.controller.TransitionTo(game.GameStatePlaying)
			g.showGame()
		})
	}

	// Setup craft button
	if craftBtn, ok := doc.GetElementById("craftBtn"); ok {
		craftBtn.UI.ToButton().Base().AddEvent(ui.EventTypeClick, func() {
			g.performCraft()
		})
	}

	g.currentDoc = doc
}

// showFallbackCrafting shows a simple crafting UI when HTML fails
func (g *TesselBoxGame) showFallbackCrafting() {
	panel := uiManager.Add().ToPanel()
	panel.Init(nil, ui.ElementTypePanel)
	panel.SetColor(matrix.Color{0.1, 0.12, 0.18, 1.0})
	panel.Base().Entity().Activate()

	title := uiManager.Add().ToLabel()
	title.Init("Crafting")
	title.SetColor(matrix.ColorWhite())
	title.SetFontSize(24)
	title.Base().Entity().Activate()
	panel.AddChild(title.Base())

	info := uiManager.Add().ToLabel()
	info.Init("Press C to craft a wooden pickaxe (requires 3 wood + 2 wood)")
	info.SetColor(matrix.Color{0.7, 0.7, 0.7, 1.0})
	info.SetFontSize(14)
	info.Base().Entity().Activate()
	panel.AddChild(info.Base())

	btnBack := uiManager.Add().ToButton()
	btnBack.Init(nil, "Close (ESC)")
	btnBack.SetColor(matrix.Color{0.5, 0.5, 0.5, 1.0})
	btnBack.Base().Entity().Activate()
	btnBack.Base().AddEvent(ui.EventTypeClick, func() {
		g.controller.TransitionTo(game.GameStatePlaying)
		g.showGame()
	})
	panel.AddChild(btnBack.Base())
}

// performCraft attempts to craft using items in the crafting grid
func (g *TesselBoxGame) performCraft() {
	player := g.controller.GetPlayer()
	if player == nil {
		return
	}

	// For now, craft a wooden pickaxe if player has wood
	inv := player.GetInventory()
	if inv.HasItem("wood", 5) {
		// Remove ingredients
		inv.RemoveItem("wood", 5)

		// Add result
		itemRegistry := crafting.GetGlobalItemRegistry()
		if pickaxe, exists := itemRegistry.GetItem("wooden_pickaxe"); exists {
			stack := crafting.NewItemStack(pickaxe, 1)
			if err := inv.AddItem(stack); err != nil {
				// Inventory full
			}
		}
	}
}

// registerStateCallbacks registers callbacks for state changes
func (g *TesselBoxGame) registerStateCallbacks() {
	// State change callbacks can be added here
}

// initiateGitHubOAuth starts the GitHub OAuth flow
func (g *TesselBoxGame) initiateGitHubOAuth() {
	if g.oauthConfig == nil {
		println("DEBUG: OAuth not configured")
		g.showLoginError("GitHub OAuth is not configured. Please set up environment variables.")
		return
	}

	// Start OAuth server in a goroutine
	go func() {
		authServer := oauth.NewAuthServer(g.oauthConfig)
		if err := authServer.StartServer(); err != nil {
			println("DEBUG: OAuth server failed to start:", err.Error())
		}
	}()

	// Open browser for OAuth
	authURL := fmt.Sprintf("http://%s:%s/auth/github/login", g.oauthConfig.ServerHost, g.oauthConfig.ServerPort)
	println("DEBUG: Opening OAuth URL:", authURL)

	// In a real implementation, you would open the system browser
	// For now, we'll just transition to main menu as a fallback
	g.transitionToMainMenu()
}

// showLoginError displays an error message on the login screen
func (g *TesselBoxGame) showLoginError(message string) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	if g.currentDoc != nil {
		// Try to get the status message element and set error text
		// This is a simplified implementation - in production you'd have proper UI element handling
		println("DEBUG: Login error:", message)
	}
}

// createFallbackLoginScreen creates a basic login UI programmatically
func (g *TesselBoxGame) createFallbackLoginScreen() {
	panel := uiManager.Add().ToPanel()
	panel.Init(nil, ui.ElementTypePanel)
	panel.SetColor(matrix.Color{0.1, 0.12, 0.18, 1.0})
	panel.Base().Entity().Activate()

	titleLabel := uiManager.Add().ToLabel()
	titleLabel.Init("TesselBox v" + Version)
	titleLabel.SetColor(matrix.ColorWhite())
	titleLabel.SetFontSize(24)
	titleLabel.Base().Entity().Activate()
	panel.AddChild(titleLabel.Base())

	btn := uiManager.Add().ToButton()
	btn.Init(nil, "Start Game")
	btn.SetColor(matrix.Color{0.91, 0.27, 0.38, 1.0})
	btn.Base().Entity().Activate()
	btn.Base().AddEvent(ui.EventTypeClick, func() {
		g.transitionToMainMenu()
	})
	panel.AddChild(btn.Base())
}

// Update handles game logic updates
func (g *TesselBoxGame) update(deltaTime float64) {
	if g.controller != nil {
		g.controller.Update()
	}
}

// getGame returns the game instance for bootstrap
func getGame() bootstrap.GameInterface {
	return &TesselBoxGame{}
}

func main() {
	bootstrap.Main(getGame(), nil)
}
