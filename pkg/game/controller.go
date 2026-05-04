package game

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/tesselstudio/TesselBox/pkg/audio"
	"github.com/tesselstudio/TesselBox/pkg/effects"
	"github.com/tesselstudio/TesselBox/pkg/network"
	"github.com/tesselstudio/TesselBox/pkg/player"
	"github.com/tesselstudio/TesselBox/pkg/ui"
	"github.com/tesselstudio/TesselBox/pkg/world"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
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

// Controller manages the game state and integrates all systems
type Controller struct {
	mu sync.RWMutex

	// Engine reference
	host *engine.Host

	// World
	world *world.World

	// Player
	player *player.Player

	// Game state
	state GameState

	// UI
	hud *ui.HUD

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

// ChunkRenderer handles rendering of a chunk
type ChunkRenderer struct {
	drawing *rendering.Drawing
	mesh    *world.ChunkMesh
}

// NewController creates a new game controller
func NewController(host *engine.Host) *Controller {
	return &Controller{
		host:             host,
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
				c.world.SetSpawnPoint(matrix.NewVec3(info.SpawnX, info.SpawnY, info.SpawnZ))
			}
		} else {
			// Save initial world info for new world
			info := c.world.GetInfo()
			saveManager.SaveWorldInfo(info)
		}
	}

	// Create player
	c.player = player.NewPlayer(c.host, c.world)

	// Load player data if it exists
	if worldExists {
		if err := player.LoadPlayer(c.player, worldName); err != nil {
			fmt.Printf("Warning: failed to load player data: %v\n", err)
			// Fall back to spawn point
			spawn := c.world.GetSpawnPoint()
			safeY := c.world.GetSafeSpawnHeight(int(spawn.X()), int(spawn.Z()))
			c.player.SetPosition(matrix.NewVec3(spawn.X(), float32(safeY), spawn.Z()))
		}
	} else {
		// Set initial position to safe spawn height for new world
		spawn := c.world.GetSpawnPoint()
		safeY := c.world.GetSafeSpawnHeight(int(spawn.X()), int(spawn.Z()))
		c.player.SetPosition(matrix.NewVec3(spawn.X(), float32(safeY), spawn.Z()))
	}

	// Create HUD
	c.hud = ui.NewHUD(1920, 1080)
	c.hud.SetPlayerStats(c.player.GetStats())
	c.hud.SetInventory(c.player.GetInventory())

	// Initialize effects and audio
	c.effectManager = effects.NewEffectManager(c.host)
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
	c.setupInput()
}

// setupInput configures input handlers
func (c *Controller) setupInput() {
	// Keyboard input
	c.host.Window.Keyboard.AddKeyCallback(func(keyId int, keyState hid.KeyState) {
		c.handleKeyInput(keyId, keyState)
	})

	// Input is polled in Update() rather than using callbacks
}

// handleKeyInput handles keyboard input
func (c *Controller) handleKeyInput(keyId int, keyState hid.KeyState) {
	c.mu.Lock()
	defer c.mu.Unlock()

	pressed := keyState == hid.KeyStateDown

	switch keyId {
	case hid.KeyboardKeyW:
		c.input.Forward = pressed
	case hid.KeyboardKeyS:
		c.input.Backward = pressed
	case hid.KeyboardKeyA:
		c.input.Left = pressed
	case hid.KeyboardKeyD:
		c.input.Right = pressed
	case hid.KeyboardKeySpace:
		c.input.Jump = pressed
		if pressed && c.state == GameStatePlaying {
			c.player.Jump()
		}
	case hid.KeyboardKeyLeftShift, hid.KeyboardKeyRightShift:
		c.input.Sprint = pressed
	case hid.KeyboardKeyLeftCtrl, hid.KeyboardKeyRightCtrl:
		c.input.Sneak = pressed
		c.player.SetSneaking(pressed)
	case hid.KeyboardKeyE:
		if pressed {
			c.toggleInventory()
		}
	case hid.KeyboardKeyEscape:
		if pressed {
			c.togglePause()
		}
	case hid.KeyboardKeyF3:
		if pressed {
			c.input.Debug = !c.input.Debug
			if c.hud != nil {
				c.hud.ToggleDebug()
			}
		}
	case hid.KeyboardKeyF5:
		if pressed {
			// Quick save
			c.saveGame()
		}
	}

	// Hotbar keys (1-9)
	if keyId >= hid.KeyboardKey1 && keyId <= hid.KeyboardKey9 {
		slot := keyId - hid.KeyboardKey1
		c.player.SetHotbarSlot(slot)
		if c.hud != nil {
			c.hud.GetHotbar().SelectSlot(slot)
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
	c.player.SetRotation(matrix.NewVec3(c.cameraPitch, c.cameraYaw, 0))
}

// handleMouseButton handles mouse button input
func (c *Controller) handleMouseButton(buttonId int, buttonState int) {
	if c.state != GameStatePlaying {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	pressed := buttonState == hid.MousePress

	switch buttonId {
	case hid.MouseButtonLeft:
		c.input.Attack = pressed
	case hid.MouseButtonRight:
		c.input.Use = pressed
	}
}

// pollMouseInput polls mouse state for clicks
func (c *Controller) pollMouseInput() {
	if c.state != GameStatePlaying {
		return
	}

	mouse := c.host.Window.Mouse

	// Left click - break block
	if mouse.Pressed(hid.MouseButtonLeft) {
		c.player.BreakBlock()
	}

	// Right click - place block
	if mouse.Pressed(hid.MouseButtonRight) {
		c.player.PlaceBlock()
	}
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
		c.world.Update(deltaTime, playerPos)

		// Update HUD
		if c.hud != nil {
			c.hud.Update(float32(deltaTime))
		}

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

// updateCamera updates the camera position
func (c *Controller) updateCamera(playerPos matrix.Vec3) {
	// Camera follows player at eye level
	eyeHeight := float32(1.6)
	cameraPos := matrix.NewVec3(playerPos.X(), playerPos.Y()+eyeHeight, playerPos.Z())

	// Update engine camera
	c.host.Cameras.Primary.Camera.SetPosition(cameraPos)

	// Calculate forward direction from rotation
	pitchRad := float64(c.cameraPitch) * math.Pi / 180.0
	yawRad := float64(c.cameraYaw) * math.Pi / 180.0

	forward := matrix.NewVec3(
		float32(-math.Sin(yawRad)*math.Cos(pitchRad)),
		float32(math.Sin(pitchRad)),
		float32(-math.Cos(yawRad)*math.Cos(pitchRad)),
	)

	// Set camera look target
	lookTarget := matrix.NewVec3(
		cameraPos.X()+forward.X(),
		cameraPos.Y()+forward.Y(),
		cameraPos.Z()+forward.Z(),
	)
	c.host.Cameras.Primary.Camera.SetLookAt(lookTarget)
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
		renderer, exists := c.chunkRenderers[coord]
		if exists && renderer.drawing != nil {
			// Drawing is already registered with the host
		}
	}
}

// updateChunkRenderer creates or updates a chunk renderer with Kaiju Engine integration
func (c *Controller) updateChunkRenderer(coord world.ChunkCoord, mesh *world.ChunkMesh) {
	renderer, exists := c.chunkRenderers[coord]
	if !exists {
		// Create new renderer with Kaiju Engine drawing
		renderer = &ChunkRenderer{}
		c.chunkRenderers[coord] = renderer

		// Create Kaiju Engine drawing for this chunk
		if len(mesh.Vertices) > 0 {
			drawing := c.createChunkDrawing(coord, mesh)
			renderer.drawing = drawing
		}
	} else {
		// Update existing renderer
		// Remove old drawing if exists
		if renderer.drawing != nil {
			// Drawing will be garbage collected when no longer referenced
			renderer.drawing = nil
		}

		// Create new drawing with updated mesh
		if len(mesh.Vertices) > 0 {
			drawing := c.createChunkDrawing(coord, mesh)
			renderer.drawing = drawing
		}
	}
}

// createChunkDrawing creates a Kaiju Engine Drawing from chunk mesh data
func (c *Controller) createChunkDrawing(coord world.ChunkCoord, mesh *world.ChunkMesh) *rendering.Drawing {
	if len(mesh.Vertices) == 0 || len(mesh.Indices) == 0 {
		return nil
	}

	// Create entity for this chunk
	entity := engine.NewEntity(c.host.WorkGroup())

	// Position entity at chunk origin
	chunkWorldX := float32(coord.X * world.ChunkSize)
	chunkWorldZ := float32(coord.Z * world.ChunkSize)
	entity.Transform.SetPosition(matrix.NewVec3(chunkWorldX, 0, chunkWorldZ))

	// Create the mesh
	kaijuMesh := c.createKaijuMesh(mesh)
	if kaijuMesh == nil {
		return nil
	}

	// Create shader data (basic shader with color support)
	sd := shader_data_registry.Create("basic")
	if sd != nil {
		if basic, ok := sd.(*shader_data_registry.ShaderDataStandard); ok {
			basic.Color = matrix.ColorWhite()
		}
	}

	// Get basic material
	mat, err := c.host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		// Fallback: create without material
		return nil
	}

	// Create texture (white square for now)
	tex, err := c.host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	if err != nil {
		tex = nil
	}

	// Create the drawing
	draw := &rendering.Drawing{
		Material:   mat.CreateInstance([]*rendering.Texture{tex}),
		Mesh:       kaijuMesh,
		ShaderData: sd,
		Transform:  &entity.Transform,
		ViewCuller: &c.host.Cameras.Primary,
	}

	// Add to host drawings
	c.host.Drawings.AddDrawing(*draw)

	return draw
}

// createKaijuMesh converts world.ChunkMesh to Kaiju Engine mesh format
func (c *Controller) createKaijuMesh(mesh *world.ChunkMesh) *rendering.Mesh {
	if len(mesh.Vertices) == 0 || len(mesh.Indices) == 0 {
		return nil
	}

	// Convert to Kaiju Vertex format
	verts := make([]rendering.Vertex, len(mesh.Vertices))
	for i := range mesh.Vertices {
		verts[i].Position = mesh.Vertices[i]
		if len(mesh.Normals) > i {
			verts[i].Normal = mesh.Normals[i]
		} else {
			verts[i].Normal = matrix.NewVec3(0, 1, 0)
		}
		if len(mesh.UVs) > i {
			verts[i].UV0 = mesh.UVs[i]
		} else {
			verts[i].UV0 = matrix.NewVec2(0, 0)
		}
		if len(mesh.Colors) > i {
			verts[i].Color = mesh.Colors[i]
		} else {
			verts[i].Color = matrix.ColorWhite()
		}
		// Set default tangent
		verts[i].Tangent = matrix.NewVec4(1, 0, 0, 1)
	}

	// Create unique key for this chunk mesh
	key := fmt.Sprintf("chunk_%d_%d", len(mesh.Vertices), len(mesh.Indices))

	// Create Kaiju mesh
	kaijuMesh := rendering.NewMesh(key, verts, mesh.Indices)
	return kaijuMesh
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

// GetHUD returns the HUD
func (c *Controller) GetHUD() *ui.HUD {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hud
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
	c.networkManager = network.NewManager(c.host)
	c.networkManager.SetCallback(c)

	// Create remote player manager
	c.remotePlayerManager = player.NewRemotePlayerManager(c.host)

	// Connect
	return c.networkManager.Connect(address, playerName)
}

// HostServer starts a multiplayer server
func (c *Controller) HostServer(port int, playerName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isMultiplayer = true

	// Create network manager
	c.networkManager = network.NewManager(c.host)
	c.networkManager.SetCallback(c)

	// Create remote player manager
	c.remotePlayerManager = player.NewRemotePlayerManager(c.host)

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
func (c *Controller) SendBlockPlace(blockType uint8, position matrix.Vec3, rotation int) {
	c.mu.RLock()
	mgr := c.networkManager
	c.mu.RUnlock()

	if mgr != nil {
		mgr.SendBlockPlace(blockType, position, rotation)
	}
}

// SendBlockBreak sends a block break to the server
func (c *Controller) SendBlockBreak(position matrix.Vec3) {
	c.mu.RLock()
	mgr := c.networkManager
	c.mu.RUnlock()

	if mgr != nil {
		mgr.SendBlockBreak(position)
	}
}

// NetworkCallback implementations

// OnPlayerJoin handles a player joining the game
func (c *Controller) OnPlayerJoin(playerID uint32, name string, position matrix.Vec3) {
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
func (c *Controller) OnPlayerMove(playerID uint32, position, rotation matrix.Vec3) {
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
func (c *Controller) OnBlockPlace(playerID uint32, blockType uint8, position matrix.Vec3) {
	c.mu.RLock()
	w := c.world
	c.mu.RUnlock()

	if w != nil {
		// Place the block in our world
		x := int(position.X())
		y := int(position.Y())
		z := int(position.Z())
		w.SetBlock(x, y, z, world.BlockData{ID: world.BlockID(blockType)})
	}
}

// OnBlockBreak handles a remote block break
func (c *Controller) OnBlockBreak(playerID uint32, position matrix.Vec3) {
	c.mu.RLock()
	w := c.world
	c.mu.RUnlock()

	if w != nil {
		// Break the block (set to air)
		x := int(position.X())
		y := int(position.Y())
		z := int(position.Z())
		w.SetBlock(x, y, z, world.BlockData{ID: world.BlockID(0)})
	}
}

// OnChatMessage handles a chat message
func (c *Controller) OnChatMessage(playerID uint32, message string) {
	fmt.Printf("[Player %d]: %s\n", playerID, message)
	// TODO: Display in chat UI
}
