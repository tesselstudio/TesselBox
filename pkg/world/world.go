package world

import (
	"sync"

	"kaijuengine.com/matrix"
)

// World represents the game world containing all chunks and systems
type World struct {
	mu sync.RWMutex

	// World name/identifier
	Name string
	Seed int64

	// Chunk management
	chunkManager *ChunkManager

	// Terrain generation
	generator *WorldGenerator

	// Save/Load system
	saveManager *SaveManager

	// Time of day (0-24000, like Minecraft)
	timeOfDay int

	// World spawn point
	spawnPoint matrix.Vec3

	// Running state
	running bool
}

// NewWorld creates a new world with the given name and seed
func NewWorld(name string, seed int64) *World {
	w := &World{
		Name:        name,
		Seed:        seed,
		timeOfDay:   6000, // Start at noon
		spawnPoint:  matrix.NewVec3(0, 70, 0),
		running:     true,
	}

	// Create generator
	w.generator = NewWorldGenerator(seed)

	// Create chunk manager
	w.chunkManager = NewChunkManager(w, 8) // 8 chunk view distance

	return w
}

// GetChunkManager returns the world's chunk manager
func (w *World) GetChunkManager() *ChunkManager {
	return w.chunkManager
}

// GetGenerator returns the world's terrain generator
func (w *World) GetGenerator() *WorldGenerator {
	return w.generator
}

// GetSaveManager returns the world's save manager
func (w *World) GetSaveManager() *SaveManager {
	return w.saveManager
}

// SetSaveManager sets the save manager for this world
func (w *World) SetSaveManager(sm *SaveManager) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.saveManager = sm
}

// Update updates the world state (called every frame)
func (w *World) Update(deltaTime float64, playerPos matrix.Vec3) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	// Update chunk manager based on player position
	w.chunkManager.Update(playerPos)

	// Update time of day (20 minutes per day)
	w.timeOfDay += int(deltaTime * 20) // 20 ticks per second
	if w.timeOfDay >= 24000 {
		w.timeOfDay = 0
	}
}

// GetTimeOfDay returns the current time of day (0-24000)
func (w *World) GetTimeOfDay() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.timeOfDay
}

// SetTimeOfDay sets the time of day
func (w *World) SetTimeOfDay(time int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.timeOfDay = time % 24000
}

// GetDayTime returns the time as a float 0.0-1.0
func (w *World) GetDayTime() float32 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return float32(w.timeOfDay) / 24000.0
}

// IsDaytime returns true if it's daytime (6:00 - 18:00)
func (w *World) IsDaytime() bool {
	time := w.GetTimeOfDay()
	return time >= 0 && time < 12000
}

// GetSpawnPoint returns the world spawn point
func (w *World) GetSpawnPoint() matrix.Vec3 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.spawnPoint
}

// SetSpawnPoint sets the world spawn point
func (w *World) SetSpawnPoint(point matrix.Vec3) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.spawnPoint = point
}

// GetBlock returns the block at world coordinates
func (w *World) GetBlock(x, y, z int) BlockData {
	return w.chunkManager.GetBlock(x, y, z)
}

// SetBlock sets a block at world coordinates
func (w *World) SetBlock(x, y, z int, block BlockData) {
	w.chunkManager.SetBlock(x, y, z, block)
}

// GetHeightmap returns the ground height at x,z
func (w *World) GetHeightmap(x, z int) int {
	chunk := w.chunkManager.GetChunkByWorld(x, z)
	if chunk == nil {
		return 0
	}

	_, localX, _, localZ := WorldToLocal(x, 0, z)
	return int(chunk.GetHeightmap(localX, localZ))
}

// GetSafeSpawnHeight returns a safe Y height for spawning at x,z
func (w *World) GetSafeSpawnHeight(x, z int) int {
	height := w.GetHeightmap(x, z)
	if height <= 0 {
		return 70 // Default height
	}
	return int(height) + 2 // Spawn 2 blocks above ground
}

// Stop stops the world update loop
func (w *World) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.running = false

	// Save all modified chunks
	if w.saveManager != nil {
		chunks := w.chunkManager.GetLoadedChunks()
		for _, chunk := range chunks {
			if chunk.IsModified() {
				w.saveManager.SaveChunk(chunk)
			}
		}
	}
}

// IsRunning returns true if the world is running
func (w *World) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// GetLoadedChunkCount returns the number of loaded chunks
func (w *World) GetLoadedChunkCount() int {
	return w.chunkManager.GetStats().LoadedChunks
}

// WorldInfo contains metadata about a world
type WorldInfo struct {
	Name       string
	Seed       int64
	LastPlayed int64
	GameTime   int
	SpawnX     float32
	SpawnY     float32
	SpawnZ     float32
}

// GetInfo returns world metadata
func (w *World) GetInfo() WorldInfo {
	w.mu.RLock()
	defer w.mu.RUnlock()

	spawn := w.spawnPoint
	return WorldInfo{
		Name:       w.Name,
		Seed:       w.Seed,
		GameTime:   w.timeOfDay,
		SpawnX:     spawn.X(),
		SpawnY:     spawn.Y(),
		SpawnZ:     spawn.Z(),
	}
}
