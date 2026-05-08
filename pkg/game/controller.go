package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/tesselstudio/TesselBox/pkg/audio"
	"github.com/tesselstudio/TesselBox/pkg/effects"
	"github.com/tesselstudio/TesselBox/pkg/network"
	"github.com/tesselstudio/TesselBox/pkg/player"
	"github.com/tesselstudio/TesselBox/pkg/types"
	"github.com/tesselstudio/TesselBox/pkg/world"
)

// AutoSaveInterval is the time between automatic saves
const AutoSaveInterval = 5 * time.Minute

// Note: This controller bridges the world simulation with Kaiju Engine rendering

// GameState represents the current state of the game
type GameState int

const (
	GameStateLogin GameState = iota
	GameStateMainMenu
	GameStateWorldSelect
	GameStateMultiplayer
	GameStateLoading
	GameStatePlaying
	GameStatePaused
	GameStateInventory
	GameStateCrafting
	GameStateSettings
)

// Controller manages the game state and logic
type Controller struct {
	mu sync.RWMutex

	// World
	world *world.World

	// Player
	player *player.Player

	// Game state
	state GameState

	// UI (placeholder for pure OpenGL implementation)
	// hud interface will be implemented with pure OpenGL

	// Input state
	input *InputState

	// Camera control
	cameraYaw        float32
	cameraPitch      float32
	mouseSensitivity float32

	// Chunk renderers
	chunkRenderers map[world.ChunkCoord]*ChunkRenderer

	// Timing
	lastUpdate time.Time

	// View distance
	viewDistance int

	// Effects and Audio
	effectManager *effects.EffectManager
	audioManager  *audio.Manager

	// Auto-save
	lastSaveTime time.Time

	// Multiplayer
	isMultiplayer       bool
	networkManager      *network.Manager
	remotePlayerManager *player.RemotePlayerManager
}

// InputState tracks current input state
type InputState struct {
	Forward   bool
	Backward  bool
	Left      bool
	Right     bool
	Jump      bool
	Sprint    bool
	Sneak     bool
	Attack    bool
	Use       bool
	Inventory bool
	Pause     bool
	Debug     bool

	MouseDX float32
	MouseDY float32
}

// ChunkRenderer handles rendering of a chunk (placeholder for pure OpenGL implementation)
type ChunkRenderer struct {
	mesh *world.ChunkMesh
}

// NewController creates a new game controller
func NewController() *Controller {
	return &Controller{
		state:            GameStateLogin,
		input:            &InputState{},
		chunkRenderers:   make(map[world.ChunkCoord]*ChunkRenderer),
		mouseSensitivity: 0.1,
		lastUpdate:       time.Now(),
		viewDistance:     8,
	}
}

// StartWorld starts a game world (new or existing)
func (c *Controller) StartWorld(worldName string, seed int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if this is an existing world
	worldExists := world.WorldExists(worldName)

	// Create or load world
	c.world = world.NewWorld(worldName, seed)

	// Initialize save manager
	saveManager, err := world.NewSaveManager(worldName)
	if err != nil {
		fmt.Printf("Warning: failed to create save manager: %v\n", err)
	} else {
		c.world.SetSaveManager(saveManager)

		// Load world info if it exists
		if worldExists {
			if info, err := saveManager.LoadWorldInfo(); err == nil {
				c.world.Seed = info.Seed
				c.world.SetTimeOfDay(info.GameTime)
				spawn := world.Vec3{X: info.SpawnX, Y: info.SpawnY, Z: info.SpawnZ}
				c.world.SetSpawnPoint(spawn)
			}
		} else {
			// Save initial world info for new world
			info := c.world.GetInfo()
			saveManager.SaveWorldInfo(info)
		}
	}

	// Create effect manager
	c.effectManager = effects.NewEffectManager()

	// Create player
	c.player = player.NewPlayer(c.world)

	// Load player data if it exists
	if worldExists {
		if err := player.LoadPlayer(c.player, worldName); err != nil {
			fmt.Printf("Warning: failed to load player data: %v\n", err)
			// Fall back to spawn point
			spawn := c.world.GetSpawnPoint()
			safeY := c.world.GetSafeSpawnHeight(int(spawn.X), int(spawn.Z))
			playerPos := types.NewVec3(spawn.X, float32(safeY), spawn.Z)
			c.player.SetPosition(playerPos)
		}
	} else {
		// Set initial position to safe spawn height for new world
		spawn := c.world.GetSpawnPoint()
		safeY := c.world.GetSafeSpawnHeight(int(spawn.X), int(spawn.Z))
		playerPos := types.NewVec3(spawn.X, float32(safeY), spawn.Z)
		c.player.SetPosition(playerPos)
	}

	// HUD will be implemented with pure OpenGL
	// For now, this is a placeholder

	// Initialize effects and audio
	c.effectManager = effects.NewEffectManager()
	c.player.SetEffectManager(c.effectManager)
	c.audioManager = audio.GetManager()
	if err := c.audioManager.Initialize(); err != nil {
		fmt.Printf("Warning: failed to initialize audio: %v\n", err)
	}

	// Start background music
	c.audioManager.PlayMusic("exploration")

	// Initialize auto-save timer
	c.lastSaveTime = time.Now()

	// Set game state to playing
	c.state = GameStatePlaying

	// Setup input
	fmt.Println("Input system initialized (placeholder)")
}

// setupInput configures input handlers
func (c *Controller) setupInput() {
	// Input handling will be implemented with pure OpenGL/Fyne
	// For now, this is a placeholder
}

// handleKeyInput handles keyboard input (placeholder for pure OpenGL implementation)
func (c *Controller) handleKeyInput(keyId int, keyState int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	pressed := keyState == 1 // Simplified key state

	// Basic input handling (simplified for pure OpenGL)
	switch keyId {
	case 87: // W key
		c.input.Forward = pressed
	case 83: // S key
		c.input.Backward = pressed
	case 65: // A key
		c.input.Left = pressed
	case 68: // D key
		c.input.Right = pressed
	case 32: // Space key
		c.input.Jump = pressed
		if pressed && c.state == GameStatePlaying {
			c.player.Jump()
		}
	}
}

// handleMouseMove handles mouse movement
func (c *Controller) handleMouseMove(x, y float32) {
	if c.state != GameStatePlaying {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Update camera rotation
	c.cameraYaw -= x * c.mouseSensitivity
	c.cameraPitch -= y * c.mouseSensitivity

	// Clamp pitch to prevent flipping
	if c.cameraPitch > 89.0 {
		c.cameraPitch = 89.0
	} else if c.cameraPitch < -89.0 {
		c.cameraPitch = -89.0
	}

	// Update player rotation
	c.player.SetRotation(types.NewVec3(c.cameraPitch, c.cameraYaw, 0))
}

// handleMouseInput handles mouse button input (placeholder for pure OpenGL implementation)
func (c *Controller) handleMouseInput(buttonId int, buttonState int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	pressed := buttonState == 1 // Simplified button state

	switch buttonId {
	case 0: // Left mouse button
		c.input.Attack = pressed
	case 1: // Right mouse button
		c.input.Use = pressed
	}
}

// pollMouseInput polls mouse state for clicks (placeholder for pure OpenGL implementation)
func (c *Controller) pollMouseInput() {
	if c.state != GameStatePlaying {
		return
	}

	// Mouse input will be implemented with pure OpenGL
	// For now, this is a placeholder
}

// toggleInventory toggles inventory screen
func (c *Controller) toggleInventory() {
	if c.state == GameStatePlaying {
		c.state = GameStateInventory
		// Show inventory UI
	} else if c.state == GameStateInventory {
		c.state = GameStatePlaying
		// Hide inventory UI
	}
}

// togglePause toggles pause menu
func (c *Controller) togglePause() {
	if c.state == GameStatePlaying {
		c.state = GameStatePaused
	} else if c.state == GameStatePaused {
		c.state = GameStatePlaying
	}
}

// saveGame saves the current game state
func (c *Controller) saveGame() {
	if c.world == nil {
		return
	}

	saveManager := c.world.GetSaveManager()
	if saveManager == nil {
		return
	}

	// Save world info with timestamp
	info := c.world.GetInfo()
	info.LastPlayed = time.Now().Unix()
	saveManager.SaveWorldInfo(info)

	// Save player data
	if c.player != nil {
		if err := player.SavePlayer(c.player, c.world.Name); err != nil {
			fmt.Printf("Failed to save player: %v\n", err)
		}
	}

	// Save all modified chunks
	chunks := c.world.GetChunkManager().GetLoadedChunks()
	savedCount := 0
	for _, chunk := range chunks {
		if chunk.IsModified() {
			saveManager.SaveChunk(chunk)
			savedCount++
		}
	}

	fmt.Printf("Saved world '%s' (%d chunks)\n", c.world.Name, savedCount)
}

// Update updates the game state
func (c *Controller) Update() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	deltaTime := now.Sub(c.lastUpdate).Seconds()
	c.lastUpdate = now

	if c.state != GameStatePlaying {
		return
	}

	// Update player
	if c.player != nil {
		// Calculate movement input
		forward := float32(0)
		right := float32(0)

		if c.input.Forward {
			forward += 1
		}
		if c.input.Backward {
			forward -= 1
		}
		if c.input.Left {
			right -= 1
		}
		if c.input.Right {
			right += 1
		}

		c.player.Move(forward, right, c.input.Sprint)
		c.player.Update(deltaTime)

		// Update world around player
		playerPos := c.player.GetPosition()
		worldPos := world.Vec3{X: playerPos.X, Y: playerPos.Y, Z: playerPos.Z}
		c.world.Update(deltaTime, worldPos)

		// HUD updates will be implemented with pure OpenGL
		// For now, this is a placeholder

		// Update camera to follow player
		c.updateCamera(playerPos)

		// Render chunks
		c.renderChunks()

		// Poll mouse input
		c.pollMouseInput()
	}

	// Update effects (particles, etc.)
	if c.effectManager != nil {
		c.effectManager.Update(float32(deltaTime))
	}

	// Update audio system
	if c.audioManager != nil {
		c.audioManager.Update()
	}

	// Update network manager (process pending updates)
	if c.networkManager != nil {
		c.networkManager.Update()
	}

	// Update remote players
	if c.remotePlayerManager != nil {
		c.remotePlayerManager.Update(float32(deltaTime))
	}

	// Auto-save check (only in single player)
	if !c.isMultiplayer && time.Since(c.lastSaveTime) > AutoSaveInterval {
		c.autoSave()
	}
}

// autoSave performs an automatic save
func (c *Controller) autoSave() {
	if c.world == nil || c.player == nil {
		return
	}

	saveManager := c.world.GetSaveManager()
	if saveManager == nil {
		return
	}

	// Save world info
	info := c.world.GetInfo()
	info.LastPlayed = time.Now().Unix()
	saveManager.SaveWorldInfo(info)

	// Save player
	if err := player.SavePlayer(c.player, c.world.Name); err != nil {
		fmt.Printf("Auto-save failed for player: %v\n", err)
	}

	// Save modified chunks
	chunks := c.world.GetChunkManager().GetLoadedChunks()
	for _, chunk := range chunks {
		if chunk.IsModified() {
			saveManager.SaveChunk(chunk)
		}
	}

	c.lastSaveTime = time.Now()
	fmt.Printf("Auto-saved world: %s\n", c.world.Name)
}

// updateCamera updates the camera position (placeholder for pure OpenGL implementation)
func (c *Controller) updateCamera(playerPos types.Vec3) {
	// Camera updates will be implemented with pure OpenGL
	// For now, this is a placeholder
}

// renderChunks renders all visible chunks
func (c *Controller) renderChunks() {
	chunks := c.world.GetChunkManager().GetLoadedChunks()

	for coord, chunk := range chunks {
		// Check if mesh needs updating
		if chunk.IsMeshDirty() {
			mesh := chunk.GetMesh()
			if mesh != nil {
				c.updateChunkRenderer(coord, mesh)
			}
		}

		// Render chunk
		_, exists := c.chunkRenderers[coord]
		if exists {
			// Chunk rendering will be implemented with pure OpenGL
		}
	}
}

// updateChunkRenderer creates or updates a chunk renderer (placeholder for pure OpenGL implementation)
func (c *Controller) updateChunkRenderer(coord world.ChunkCoord, mesh *world.ChunkMesh) {
	renderer, exists := c.chunkRenderers[coord]
	if !exists {
		// Create new renderer
		renderer = &ChunkRenderer{}
		c.chunkRenderers[coord] = renderer
	}

	// Update mesh data
	renderer.mesh = mesh
}

// createChunkDrawing creates a chunk rendering (placeholder for pure OpenGL implementation)
func (c *Controller) createChunkDrawing(coord world.ChunkCoord, mesh *world.ChunkMesh) interface{} {
	if len(mesh.Vertices) == 0 || len(mesh.Indices) == 0 {
		return nil
	}

	// Chunk rendering will be implemented with pure OpenGL
	// For now, this is a placeholder
	return nil
}

// createKaijuMesh converts world.ChunkMesh to mesh format (placeholder for pure OpenGL implementation)
func (c *Controller) createKaijuMesh(mesh *world.ChunkMesh) interface{} {
	if len(mesh.Vertices) == 0 || len(mesh.Indices) == 0 {
		return nil
	}

	// Mesh creation will be implemented with pure OpenGL
	// For now, this is a placeholder
	return nil
}

// GetState returns the current game state
func (c *Controller) GetState() GameState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// GetWorld returns the game world
func (c *Controller) GetWorld() *world.World {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.world
}

// GetPlayer returns the player
func (c *Controller) GetPlayer() *player.Player {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.player
}

// GetHUD returns the HUD interface (placeholder for pure OpenGL implementation)
func (c *Controller) GetHUD() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// HUD will be implemented with pure OpenGL
	return nil
}

// IsPlaying returns true if game is in playing state
func (c *Controller) IsPlaying() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == GameStatePlaying
}

// Stop stops the game controller
func (c *Controller) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Save game
	c.saveGame()

	// Stop world
	if c.world != nil {
		c.world.Stop()
	}

	// Shutdown audio
	if c.audioManager != nil {
		c.audioManager.Shutdown()
	}
}

// State transition methods

// TransitionTo changes the game state with validation
func (c *Controller) TransitionTo(newState GameState) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate state transition
	if !c.isValidTransition(c.state, newState) {
		return false
	}

	c.state = newState
	return true
}

// isValidTransition checks if a state transition is valid
func (c *Controller) isValidTransition(from, to GameState) bool {
	switch from {
	case GameStateLogin:
		return to == GameStateMainMenu
	case GameStateMainMenu:
		return to == GameStateWorldSelect || to == GameStateMultiplayer || to == GameStateSettings
	case GameStateWorldSelect:
		return to == GameStateLoading || to == GameStateMainMenu
	case GameStateMultiplayer:
		return to == GameStateMainMenu || to == GameStateLoading
	case GameStateLoading:
		return to == GameStatePlaying
	case GameStatePlaying:
		return to == GameStatePaused || to == GameStateInventory || to == GameStateCrafting
	case GameStatePaused:
		return to == GameStatePlaying || to == GameStateSettings || to == GameStateMainMenu
	case GameStateInventory:
		return to == GameStatePlaying
	case GameStateCrafting:
		return to == GameStatePlaying
	case GameStateSettings:
		return to == GameStateMainMenu || to == GameStatePaused
	default:
		return false
	}
}

// GoToMainMenu transitions to main menu
func (c *Controller) GoToMainMenu() {
	c.TransitionTo(GameStateMainMenu)
}

// GoToWorldSelect transitions to world selection
func (c *Controller) GoToWorldSelect() {
	c.TransitionTo(GameStateWorldSelect)
}

// GoToMultiplayer transitions to multiplayer menu
func (c *Controller) GoToMultiplayer() {
	c.TransitionTo(GameStateMultiplayer)
}

// StartLoading transitions to loading state
func (c *Controller) StartLoading() {
	c.TransitionTo(GameStateLoading)
}

// OpenSettings transitions to settings
func (c *Controller) OpenSettings() {
	c.TransitionTo(GameStateSettings)
}

// CloseSettings returns from settings
func (c *Controller) CloseSettings() {
	if c.state == GameStateSettings {
		if c.world != nil {
			c.TransitionTo(GameStatePaused)
		} else {
			c.TransitionTo(GameStateMainMenu)
		}
	}
}

// IsInMenu returns true if in a menu state
func (c *Controller) IsInMenu() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == GameStateMainMenu || c.state == GameStateWorldSelect ||
		c.state == GameStateMultiplayer || c.state == GameStateSettings
}

// IsPaused returns true if game is paused
func (c *Controller) IsPaused() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == GameStatePaused || c.state == GameStateInventory ||
		c.state == GameStateCrafting || c.state == GameStateSettings
}

// SetNetworkClient sets the network client for multiplayer
func (c *Controller) SetNetworkClient(client interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Store client for network communication
	// TODO: Full network integration
}

// SetNetworkServer sets the network server for multiplayer hosting
func (c *Controller) SetNetworkServer(server interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Store server for multiplayer hosting
	// TODO: Full server integration
}

// ConnectToServer connects to a multiplayer server
func (c *Controller) ConnectToServer(address, playerName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isMultiplayer = true

	// Create network manager
	c.networkManager = network.NewManager()
	c.networkManager.SetCallback(c)

	// Create remote player manager
	c.remotePlayerManager = player.NewRemotePlayerManager()

	// Connect
	return c.networkManager.Connect(address, playerName)
}

// HostServer starts a multiplayer server
func (c *Controller) HostServer(port int, playerName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isMultiplayer = true

	// Create network manager
	c.networkManager = network.NewManager()
	c.networkManager.SetCallback(c)

	// Create remote player manager
	c.remotePlayerManager = player.NewRemotePlayerManager()

	// Host and connect
	return c.networkManager.HostServer(port, playerName)
}

// DisconnectMultiplayer disconnects from multiplayer
func (c *Controller) DisconnectMultiplayer() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.networkManager != nil {
		c.networkManager.Disconnect()
		c.networkManager = nil
	}

	if c.remotePlayerManager != nil {
		// Clean up remote players
		players := c.remotePlayerManager.GetAllPlayers()
		for _, player := range players {
			player.Destroy()
		}
		c.remotePlayerManager = nil
	}

	c.isMultiplayer = false
}

// IsMultiplayer returns true if in multiplayer mode
func (c *Controller) IsMultiplayer() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isMultiplayer
}

// SendBlockPlace sends a block placement to the server
func (c *Controller) SendBlockPlace(blockType uint8, position types.Vec3, rotation int) {
	c.mu.RLock()
	mgr := c.networkManager
	c.mu.RUnlock()

	if mgr != nil {
		mgr.SendBlockPlace(blockType, position, rotation)
	}
}

// SendBlockBreak sends a block break to the server
func (c *Controller) SendBlockBreak(position types.Vec3) {
	c.mu.RLock()
	mgr := c.networkManager
	c.mu.RUnlock()

	if mgr != nil {
		mgr.SendBlockBreak(position)
	}
}

// NetworkCallback implementations

// OnPlayerJoin handles a player joining the game
func (c *Controller) OnPlayerJoin(playerID uint32, name string, position types.Vec3) {
	fmt.Printf("Player joined: %s (ID: %d)\n", name, playerID)

	c.mu.Lock()
	rpm := c.remotePlayerManager
	c.mu.Unlock()

	if rpm != nil {
		rpm.CreatePlayer(playerID, name, position)
	}
}

// OnPlayerLeave handles a player leaving the game
func (c *Controller) OnPlayerLeave(playerID uint32) {
	fmt.Printf("Player left: %d\n", playerID)

	c.mu.Lock()
	rpm := c.remotePlayerManager
	c.mu.Unlock()

	if rpm != nil {
		rpm.RemovePlayer(playerID)
	}
}

// OnPlayerMove handles a player movement update
func (c *Controller) OnPlayerMove(playerID uint32, position, rotation types.Vec3) {
	c.mu.RLock()
	rpm := c.remotePlayerManager
	c.mu.RUnlock()

	if rpm != nil {
		if player := rpm.GetPlayer(playerID); player != nil {
			player.SetPosition(position)
			player.SetRotation(rotation)
		}
	}
}

// OnBlockPlace handles a remote block placement
func (c *Controller) OnBlockPlace(playerID uint32, blockType uint8, position types.Vec3) {
	c.mu.RLock()
	w := c.world
	c.mu.RUnlock()

	if w != nil {
		// Place the block in our world
		x := int(position.X)
		y := int(position.Y)
		z := int(position.Z)
		w.SetBlock(x, y, z, world.BlockData{ID: world.BlockID(blockType)})
	}
}

// OnBlockBreak handles a remote block break
func (c *Controller) OnBlockBreak(playerID uint32, position types.Vec3) {
	c.mu.RLock()
	w := c.world
	c.mu.RUnlock()

	if w != nil {
		// Break the block (set to air)
		x := int(position.X)
		y := int(position.Y)
		z := int(position.Z)
		w.SetBlock(x, y, z, world.BlockData{ID: world.BlockID(0)})
	}
}

// OnChatMessage handles a chat message
func (c *Controller) OnChatMessage(playerID uint32, message string) {
	fmt.Printf("[Player %d]: %s\n", playerID, message)
	// TODO: Display in chat UI
}
