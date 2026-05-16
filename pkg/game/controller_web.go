//go:build js && wasm
// +build js,wasm

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

// Controller manages the game state and logic (Web version without OpenGL)
type Controller struct {
	mu sync.RWMutex

	// World
	world *world.World

	// Player
	player *player.Player

	// Game state
	state GameState

	// Input state
	input *InputState

	// Camera control
	cameraYaw        float32
	cameraPitch      float32
	mouseSensitivity float32
	lastCameraSwitch time.Time

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

// NewController creates a new game controller
func NewController() *Controller {
	return &Controller{
		state:            GameStateLogin,
		input:            &InputState{},
		mouseSensitivity: 0.1,
		lastUpdate:       time.Now(),
		viewDistance:     8,
	}
}

// StartWorld starts a game world (new or existing)
func (c *Controller) StartWorld(worldName string, seed int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Println("Starting world:", worldName, "seed:", seed)

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

	// Initialize effects and audio
	c.effectManager = effects.NewEffectManager()
	c.player.SetEffectManager(c.effectManager)
	c.audioManager = audio.GetManager()
	if err := c.audioManager.Initialize(); err != nil {
		fmt.Printf("Warning: failed to initialize audio: %v\n", err)
	}

	// Initialize auto-save timer
	c.lastSaveTime = time.Now()

	// Set game state to playing
	c.state = GameStatePlaying

	// Initialize chunk loading around player
	if c.world != nil && c.world.GetChunkManager() != nil {
		c.world.GetChunkManager().InitializeChunkLoading(c.player.GetPosition())
	}

	fmt.Println("World started successfully")
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

		// Update player rotation from camera
		c.player.SetRotation(types.NewVec3(c.cameraPitch, c.cameraYaw, 0))
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

// HandleMouseMove handles mouse movement
func (c *Controller) HandleMouseMove(x, y float32) {
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
}

// HandleKeyInput handles keyboard input
func (c *Controller) HandleKeyInput(keyCode int, keyState int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	pressed := keyState == 1

	// Basic input handling
	switch keyCode {
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
		if pressed && c.state == GameStatePlaying && c.player != nil {
			c.player.Jump()
		}
	case 27: // ESC key
		c.input.Pause = pressed
		if pressed && time.Since(c.lastCameraSwitch) > 150*time.Millisecond {
			if c.state == GameStatePlaying {
				c.state = GameStatePaused
				fmt.Println("Game Paused")
			} else if c.state == GameStatePaused {
				c.state = GameStatePlaying
				fmt.Println("Game Resumed")
			}
			c.lastCameraSwitch = time.Now()
		}
	}
}

// HandleMouseInput handles mouse button input
func (c *Controller) HandleMouseInput(buttonId int, buttonState int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	pressed := buttonState == 1

	switch buttonId {
	case 0: // Left mouse button
		c.input.Attack = pressed
	case 1: // Right mouse button
		c.input.Use = pressed
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

	fmt.Println("Game stopped")
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

// IsPlaying returns true if game is in playing state
func (c *Controller) IsPlaying() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == GameStatePlaying
}
