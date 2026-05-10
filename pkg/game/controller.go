package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tesselstudio/TesselBox/pkg/audio"
	"github.com/tesselstudio/TesselBox/pkg/effects"
	"github.com/tesselstudio/TesselBox/pkg/network"
	"github.com/tesselstudio/TesselBox/pkg/opengl"
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

// CameraManagerInterface defines the interface for camera management
type CameraManagerInterface interface {
	UpdateFromPlayer(playerPos types.Vec3, playerYaw, playerPitch float32)
	GetViewMatrix() mgl32.Mat4
	GetProjectionMatrix() mgl32.Mat4
	GetPosition() mgl32.Vec3
	GetFront() mgl32.Vec3
	GetRight() mgl32.Vec3
	GetUp() mgl32.Vec3
	SetAspectRatio(width, height int)
	SetFOV(fov float32)
	SetClipPlanes(near, far float32)
	CycleMode()
	GetMode() interface{ String() string }
}

// CameraModeInterface defines the interface for camera modes
type CameraModeInterface interface {
	String() string
}

// Controller manages the game state and logic
type Controller struct {
	mu sync.RWMutex

	// World
	world *world.World

	// Player
	player *player.Player

	// Game state
	state GameState

	// HUD
	hud *HUD

	// Input state
	input *InputState

	// Camera control
	cameraManager    CameraManagerInterface
	cameraYaw        float32
	cameraPitch      float32
	mouseSensitivity float32
	lastCameraSwitch time.Time

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
	Forward      bool
	Backward     bool
	Left         bool
	Right        bool
	Jump         bool
	Sprint       bool
	Sneak        bool
	Attack       bool
	Use          bool
	Inventory    bool
	Pause        bool
	Debug        bool
	Drop         bool
	Crafting     bool
	CameraSwitch bool
	MenuReturn   bool

	MouseDX    float32
	MouseDY    float32
	HotbarSlot int
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
		cameraManager:    nil, // Will be set externally
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

	// Initialize HUD
	c.hud = NewHUD(c.player)

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

	// Initialize chunk loading around player
	if c.world != nil && c.world.GetChunkManager() != nil {
		c.world.GetChunkManager().InitializeChunkLoading(c.player.GetPosition())
	}

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

// HandleMouseMove is the public method for handling mouse movement
func (c *Controller) HandleMouseMove(x, y float32) {
	c.handleMouseMove(x, y)
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

// togglePause toggles pause menu, returns true if already paused (should exit to menu)
func (c *Controller) togglePause() bool {
	if c.state == GameStatePlaying {
		c.state = GameStatePaused
		return false
	} else if c.state == GameStatePaused {
		c.state = GameStatePlaying
		return true
	}
	return false
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

	// Handle pause input first (works in both playing and paused states)
	// ESC while playing -> pause, ESC while paused -> return to menu
	if c.input.Pause && time.Since(c.lastCameraSwitch) > 150*time.Millisecond {
		if c.state == GameStatePlaying {
			c.state = GameStatePaused
			fmt.Println("⏸️  Game Paused")
		} else if c.state == GameStatePaused {
			// ESC pressed while paused - return to main menu
			c.ReturnToMainMenuFromGame()
		}
		c.lastCameraSwitch = time.Now()
	}

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

		// Handle camera switching with debouncing
		if c.input.CameraSwitch && time.Since(c.lastCameraSwitch) > 200*time.Millisecond {
			c.cameraManager.CycleMode()
			c.lastCameraSwitch = time.Now()
		}

		// Update world around player
		playerPos := c.player.GetPosition()
		worldPos := world.Vec3{X: playerPos.X, Y: playerPos.Y, Z: playerPos.Z}
		c.world.Update(deltaTime, worldPos)

		// Update HUD
		c.hud.Update()

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

// updateCamera updates the camera position using the camera manager
func (c *Controller) updateCamera(playerPos types.Vec3) {
	if c.cameraManager != nil {
		// Update camera manager with player position and rotation
		c.cameraManager.UpdateFromPlayer(playerPos, c.cameraYaw, c.cameraPitch)
	}
}

// SetCameraManager sets the camera manager (called from launch_game.go)
func (c *Controller) SetCameraManager(manager CameraManagerInterface) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cameraManager = manager
}

// GetCameraManager returns the camera manager interface
func (c *Controller) GetCameraManager() CameraManagerInterface {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cameraManager
}

// GetCameraMode returns the current camera mode as interface
func (c *Controller) GetCameraMode() CameraModeInterface {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cameraManager != nil {
		return c.cameraManager.GetMode()
	}
	return nil
}

// renderChunks renders all visible chunks
func (c *Controller) renderChunks() {
	chunks := c.world.GetChunkManager().GetLoadedChunks()

	for coord, chunk := range chunks {
		// Check if mesh needs updating
		var chunkMesh *world.ChunkMesh
		if chunk.IsMeshDirty() {
			chunkMesh = chunk.GetMesh()
			if chunkMesh != nil {
				c.updateChunkRenderer(coord, chunkMesh)
			}
		}

		// Render chunk using OpenGL mesh renderer
		if chunkMesh != nil {
			c.renderChunkMesh(coord, chunkMesh)
		}
	}
}

// renderChunkMesh renders a single chunk mesh using OpenGL
func (c *Controller) renderChunkMesh(coord world.ChunkCoord, chunkMesh *world.ChunkMesh) {
	if chunkMesh == nil || len(chunkMesh.Vertices) == 0 || len(chunkMesh.Indices) == 0 {
		return
	}

	// Convert world.ChunkMesh to OpenGL ChunkMeshData
	meshData := c.convertChunkMeshToOpenGL(chunkMesh)
	if meshData != nil {
		// This would be handled by the OpenGL engine's mesh renderer
		// For now, we'll use a simple direct rendering approach
		c.renderChunkDirectly(meshData)
	}
}

// convertChunkMeshToOpenGL converts world.ChunkMesh to opengl.ChunkMeshData
func (c *Controller) convertChunkMeshToOpenGL(chunkMesh *world.ChunkMesh) *opengl.ChunkMeshData {
	if chunkMesh == nil {
		return nil
	}

	// Convert types.Vec3 to float32 vertices
	vertices := make([]float32, 0)
	for _, vertex := range chunkMesh.Vertices {
		vertices = append(vertices, vertex.X, vertex.Y, vertex.Z)
	}

	return &opengl.ChunkMeshData{
		Vertices:    vertices,
		Indices:     chunkMesh.Indices,
		VertexCount: int32(len(vertices) / 3), // Assuming 3 vertices per position
		IndexCount:  int32(len(chunkMesh.Indices)),
	}
}

// renderChunkDirectly renders chunk mesh directly with OpenGL
func (c *Controller) renderChunkDirectly(meshData *opengl.ChunkMeshData) {
	if meshData == nil {
		return
	}

	// Use a simple shader for chunk rendering
	vertexShader := `
		#version 410 core
		layout (location = 0) in vec3 aPos;
		layout (location = 1) in vec3 aColor;
		uniform mat4 model;
		uniform mat4 view;
		uniform mat4 projection;
		out vec3 FragColor;
		
		void main() {
			FragColor = aColor;
			gl_Position = projection * view * model * vec4(aPos, 1.0);
		}
	`

	fragmentShader := `
		#version 410 core
		in vec3 FragColor;
		out vec4 color;
		
		void main() {
			color = vec4(FragColor, 1.0);
		}
	`

	// Create and compile shader program (simplified for this example)
	shaderProgram := c.createChunkShaderProgram(vertexShader, fragmentShader)
	if shaderProgram == 0 {
		return
	}

	defer gl.DeleteProgram(shaderProgram)

	// Create VAO and VBO for chunk
	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	// Bind and fill VBO
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(meshData.Vertices)*4, gl.Ptr(meshData.Vertices), gl.STATIC_DRAW)

	// Set vertex attributes
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0)) // Position
	gl.EnableVertexAttribArray(0)
	if len(meshData.Vertices) >= 6 { // Check if we have color data
		gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(3*4)) // Color
		gl.EnableVertexAttribArray(1)
	}

	// Set up matrices
	model := mgl32.Ident4()
	view := mgl32.LookAtV(
		mgl32.Vec3{0, 10, 5}, // Camera position
		mgl32.Vec3{0, 0, 0},  // Look at origin
		mgl32.Vec3{0, 1, 0},  // Up vector
	)
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), 800.0/600.0, 0.1, 100.0)

	modelLoc := gl.GetUniformLocation(shaderProgram, gl.Str("model\x00"))
	viewLoc := gl.GetUniformLocation(shaderProgram, gl.Str("view\x00"))
	projLoc := gl.GetUniformLocation(shaderProgram, gl.Str("projection\x00"))

	gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])
	gl.UniformMatrix4fv(viewLoc, 1, false, &view[0])
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

	// Draw the chunk
	gl.DrawArrays(gl.TRIANGLES, 0, meshData.VertexCount)

	// Cleanup
	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.DeleteVertexArrays(1, &vao)
	gl.DeleteBuffers(1, &vbo)
}

// createChunkShaderProgram creates and compiles shaders for chunk rendering
func (c *Controller) createChunkShaderProgram(vertexSource, fragmentSource string) uint32 {
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	cstr, free := gl.Strs(vertexSource)
	gl.ShaderSource(vertexShader, 1, cstr, nil)
	gl.CompileShader(vertexShader)
	free()

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	cstr, free = gl.Strs(fragmentSource)
	gl.ShaderSource(fragmentShader, 1, cstr, nil)
	gl.CompileShader(fragmentShader)
	free()

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// Check linking status
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status != gl.TRUE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		logBytes := make([]byte, logLength+1)
		gl.GetProgramInfoLog(program, logLength, nil, &logBytes[0])
		log := string(logBytes)
		fmt.Printf("Chunk shader linking error: %v\n", log)
		gl.DeleteProgram(program)
		return 0
	}

	// Clean up shaders
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program
}

// updateChunkRenderer creates or updates a chunk renderer
func (c *Controller) updateChunkRenderer(coord world.ChunkCoord, mesh *world.ChunkMesh) {
	// Store mesh reference for rendering
	// The actual rendering is handled in renderChunkMesh
}

// createChunkDrawing creates a chunk rendering (placeholder for pure OpenGL implementation)
func (c *Controller) createChunkDrawing(coord world.ChunkCoord, mesh *world.ChunkMesh) interface{} {
	// This method is replaced by renderChunkMesh
	return nil
}

// createKaijuMesh converts world.ChunkMesh to mesh format (placeholder for pure OpenGL implementation)
func (c *Controller) createKaijuMesh(mesh *world.ChunkMesh) interface{} {
	// This method is replaced by convertChunkMeshToOpenGL
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

// GetHUD returns the HUD interface
func (c *Controller) GetHUD() *HUD {
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

// Render renders the game world and UI
func (c *Controller) Render(width, height int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != GameStatePlaying {
		return
	}

	// Render HUD
	if c.hud != nil {
		c.hud.Render(width, height)
	}
}

// SetInputState updates the controller's input state
func (c *Controller) SetInputState(state *InputState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.input = state
}

// HandleJump processes jump action
func (c *Controller) HandleJump() {
	if c.player != nil && c.state == GameStatePlaying {
		c.player.Jump()
	}
}

// HandleAttack processes attack action
func (c *Controller) HandleAttack() {
	if c.player != nil && c.state == GameStatePlaying {
		// TODO: Implement block breaking
		fmt.Println("Attack action triggered")
	}
}

// HandleUse processes use action
func (c *Controller) HandleUse() {
	if c.player != nil && c.state == GameStatePlaying {
		// TODO: Implement block placing
		fmt.Println("Use action triggered")
	}
}

// HandleDropItem processes drop item action
func (c *Controller) HandleDropItem() {
	if c.player != nil && c.state == GameStatePlaying {
		// TODO: Implement item dropping
		fmt.Println("Drop item action triggered")
	}
}

// SetHotbarSlot sets the active hotbar slot
func (c *Controller) SetHotbarSlot(slot int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.input != nil {
		c.input.HotbarSlot = slot
	}
	if c.player != nil {
		c.player.SetHotbarSlot(slot)
	}
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

// ReturnToMainMenuFromGame saves game and returns to main menu
func (c *Controller) ReturnToMainMenuFromGame() {
	fmt.Println("🚪 Returning to main menu...")

	// Save game before returning to menu
	c.saveGame()

	// Transition to main menu
	c.TransitionTo(GameStateMainMenu)
	fmt.Println("✅ Game ended - returned to main menu")
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

// checkAutoSave checks if it's time to auto-save and saves if needed
func (c *Controller) checkAutoSave(deltaTime float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.world == nil || c.state != GameStatePlaying {
		return
	}

	// Update last save time
	now := time.Now()
	if now.Sub(c.lastSaveTime) >= AutoSaveInterval {
		c.saveGame()
		c.lastSaveTime = now
		fmt.Println("💾 Auto-save completed")
	}
}
