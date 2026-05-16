package world

import (
	"math"
	"sync"
	"time"

	"github.com/tesselstudio/TesselBox/pkg/blocks"
	"github.com/tesselstudio/TesselBox/pkg/types"
)

// ChunkManager manages loading, unloading, and rendering of chunks
type ChunkManager struct {
	mu sync.RWMutex

	// World reference
	world *World

	// Loaded chunks map
	chunks map[ChunkCoord]*Chunk

	// Chunk generation queue
	generateQueue chan ChunkCoord

	// Chunk mesh rebuild queue
	meshQueue chan *Chunk

	// View distance in chunks
	viewDistance int

	// Player position for chunk loading
	playerChunk ChunkCoord

	// Worker pool for chunk generation
	workers int

	// Chunk cache for recently unloaded chunks
	cache        map[ChunkCoord]*Chunk
	cacheMu      sync.RWMutex
	maxCacheSize int

	// Statistics
	stats ChunkStats
}

// ChunkStats holds chunk manager statistics
type ChunkStats struct {
	LoadedChunks    int
	GeneratedChunks int
	MeshedChunks    int
	CachedChunks    int
}

// NewChunkManager creates a new chunk manager
func NewChunkManager(world *World, viewDistance int) *ChunkManager {
	cm := &ChunkManager{
		world:         world,
		chunks:        make(map[ChunkCoord]*Chunk),
		generateQueue: make(chan ChunkCoord, 100),
		meshQueue:     make(chan *Chunk, 100),
		viewDistance:  viewDistance,
		workers:       4,
		cache:         make(map[ChunkCoord]*Chunk),
		maxCacheSize:  100,
	}

	// Start worker goroutines
	for i := 0; i < cm.workers; i++ {
		go cm.generationWorker()
	}

	// Start mesh builder goroutine
	go cm.meshWorker()

	return cm
}

// GetChunk returns a loaded chunk by coordinate
func (cm *ChunkManager) GetChunk(coord ChunkCoord) *Chunk {
	cm.mu.RLock()
	chunk, exists := cm.chunks[coord]
	cm.mu.RUnlock()

	if exists {
		return chunk
	}

	// Check cache
	cm.cacheMu.RLock()
	cachedChunk, cached := cm.cache[coord]
	cm.cacheMu.RUnlock()

	if cached {
		// Move from cache to active chunks
		cm.mu.Lock()
		cm.chunks[coord] = cachedChunk
		cm.mu.Unlock()

		cm.cacheMu.Lock()
		delete(cm.cache, coord)
		cm.cacheMu.Unlock()

		return cachedChunk
	}

	return nil
}

// GetChunkByWorld returns a chunk by world coordinates
func (cm *ChunkManager) GetChunkByWorld(worldX, worldZ int) *Chunk {
	coord, _, _, _ := WorldToLocal(worldX, 0, worldZ)
	return cm.GetChunk(coord)
}

// GetBlock returns a block at world coordinates
func (cm *ChunkManager) GetBlock(worldX, worldY, worldZ int) BlockData {
	coord, localX, localY, localZ := WorldToLocal(worldX, worldY, worldZ)

	chunk := cm.GetChunk(coord)
	if chunk == nil {
		return BlockData{ID: BlockIDAir}
	}

	return chunk.GetBlock(localX, localY, localZ)
}

// SetBlock sets a block at world coordinates
func (cm *ChunkManager) SetBlock(worldX, worldY, worldZ int, block BlockData) {
	coord, localX, localY, localZ := WorldToLocal(worldX, worldY, worldZ)

	chunk := cm.GetChunk(coord)
	if chunk == nil {
		// Create chunk if it doesn't exist
		chunk = cm.createChunk(coord)
	}

	chunk.SetBlock(localX, localY, localZ, block)

	// Mark neighbor chunks as dirty if on border
	if localX == 0 {
		neighbor := cm.GetChunk(ChunkCoord{X: coord.X - 1, Z: coord.Z})
		if neighbor != nil {
			neighbor.MarkMeshClean() // Force rebuild
		}
	}
	if localX == ChunkSize-1 {
		neighbor := cm.GetChunk(ChunkCoord{X: coord.X + 1, Z: coord.Z})
		if neighbor != nil {
			neighbor.MarkMeshClean()
		}
	}
	if localZ == 0 {
		neighbor := cm.GetChunk(ChunkCoord{X: coord.X, Z: coord.Z - 1})
		if neighbor != nil {
			neighbor.MarkMeshClean()
		}
	}
	if localZ == ChunkSize-1 {
		neighbor := cm.GetChunk(ChunkCoord{X: coord.X, Z: coord.Z + 1})
		if neighbor != nil {
			neighbor.MarkMeshClean()
		}
	}
}

// Update updates chunk loading based on player position
func (cm *ChunkManager) Update(playerPos types.Vec3) {
	// Convert player position to chunk coordinate
	playerChunkX := int(math.Floor(float64(playerPos.GetX()) / float64(ChunkSize)))
	playerChunkZ := int(math.Floor(float64(playerPos.GetZ()) / float64(ChunkSize)))
	newPlayerChunk := ChunkCoord{X: playerChunkX, Z: playerChunkZ}

	// Check if player moved to a new chunk
	if newPlayerChunk != cm.playerChunk {
		cm.playerChunk = newPlayerChunk
		cm.loadChunksAroundPlayer()
		cm.unloadDistantChunks()
	}
}

// loadChunksAroundPlayer loads chunks within view distance
func (cm *ChunkManager) loadChunksAroundPlayer() {
	dist := cm.viewDistance

	for x := -dist; x <= dist; x++ {
		for z := -dist; z <= dist; z++ {
			// Circular view distance
			if x*x+z*z > dist*dist {
				continue
			}

			coord := ChunkCoord{
				X: cm.playerChunk.X + x,
				Z: cm.playerChunk.Z + z,
			}

			cm.mu.RLock()
			_, exists := cm.chunks[coord]
			cm.mu.RUnlock()

			if !exists {
				// Queue for generation
				select {
				case cm.generateQueue <- coord:
				default:
					// Queue full, skip for now
				}
			}
		}
	}
}

// unloadDistantChunks unloads chunks outside view distance
func (cm *ChunkManager) unloadDistantChunks() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	dist := cm.viewDistance + 2 // Add buffer

	for coord, chunk := range cm.chunks {
		dx := coord.X - cm.playerChunk.X
		dz := coord.Z - cm.playerChunk.Z

		if dx*dx+dz*dz > dist*dist {
			// Save chunk if modified
			if chunk.IsModified() && cm.world.saveManager != nil {
				cm.world.saveManager.SaveChunk(chunk)
			}

			// Move to cache instead of deleting immediately
			cm.cacheMu.Lock()
			if len(cm.cache) < cm.maxCacheSize {
				cm.cache[coord] = chunk
			}
			cm.cacheMu.Unlock()

			delete(cm.chunks, coord)
		}
	}
}

// createChunk creates a new chunk and generates terrain
func (cm *ChunkManager) createChunk(coord ChunkCoord) *Chunk {
	chunk := NewChunk(coord)

	cm.mu.Lock()
	cm.chunks[coord] = chunk
	cm.mu.Unlock()

	// Generate terrain
	if cm.world.generator != nil {
		cm.world.generator.GenerateChunk(chunk)
	}

	return chunk
}

// generationWorker processes chunk generation queue
func (cm *ChunkManager) generationWorker() {
	println("🔄 Chunk generation worker started")
	for {
		select {
		case coord, ok := <-cm.generateQueue:
			if !ok {
				// Channel closed, exit worker
				return
			}

			cm.mu.RLock()
			_, exists := cm.chunks[coord]
			cm.mu.RUnlock()

			if exists {
				continue
			}

			println("🏗️ Generating chunk at", coord.X, coord.Z)
			cm.createChunk(coord)
			println("✅ Chunk generated at", coord.X, coord.Z)
		}
	}
}

// meshWorker processes chunk mesh rebuilds
func (cm *ChunkManager) meshWorker() {
	ticker := time.NewTicker(50 * time.Millisecond) // 20 mesh updates per second max
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.RebuildDirtyMeshes()
		}
	}
}

// RebuildDirtyMeshes rebuilds meshes for dirty chunks
func (cm *ChunkManager) RebuildDirtyMeshes() {
	cm.mu.RLock()
	chunks := make([]*Chunk, 0, len(cm.chunks))
	for _, chunk := range cm.chunks {
		if chunk.IsMeshDirty() {
			chunks = append(chunks, chunk)
		}
	}
	cm.mu.RUnlock()

	// Rebuild meshes
	for _, chunk := range chunks {
		mesh := cm.buildChunkMesh(chunk)
		chunk.SetMesh(mesh)

		println("🎨 Rebuilt mesh for chunk", chunk.Coord.X, chunk.Coord.Z, "with", len(mesh.Vertices), "vertices")
	}
}

// UpdateEngineWithMeshes passes chunk meshes to OpenGL engine for rendering
func (cm *ChunkManager) UpdateEngineWithMeshes(engine interface{}) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Type assert to get engine methods
	type EngineInterface interface {
		AddChunkMesh(coord ChunkCoord, vertices []float32, indices []uint32)
		RemoveChunkMesh(coord ChunkCoord)
	}

	if engineInterface, ok := engine.(EngineInterface); ok {
		for coord, chunk := range cm.chunks {
			if !chunk.IsMeshUploaded() && chunk.mesh != nil && len(chunk.mesh.InterleavedVertices) > 0 {
				engineInterface.AddChunkMesh(coord, chunk.mesh.InterleavedVertices, chunk.mesh.Indices)
				chunk.MarkMeshUploaded()
			}
		}
	}
}

// convertToFloat32Slice converts []types.Vec3 to []float32
func (cm *ChunkManager) convertToFloat32Slice(vertices []types.Vec3) []float32 {
	result := make([]float32, len(vertices)*3)
	for i, v := range vertices {
		result[i*3] = float32(v.GetX())
		result[i*3+1] = float32(v.GetY())
		result[i*3+2] = float32(v.GetZ())
	}
	return result
}

// buildChunkMesh builds the renderable mesh for a chunk
func (cm *ChunkManager) buildChunkMesh(chunk *Chunk) *ChunkMesh {
	// Use the ChunkMeshBuilder which properly handles hexagonal prisms
	builder := NewChunkMeshBuilder(chunk)
	return builder.BuildMesh()
}

// isBlockVisible checks if a block has any exposed faces
func (cm *ChunkManager) isBlockVisible(chunk *Chunk, x, y, z int) bool {
	// Check all 6 neighbors
	neighbors := [][3]int{
		{x + 1, y, z},
		{x - 1, y, z},
		{x, y + 1, z},
		{x, y - 1, z},
		{x, y, z + 1},
		{x, y, z - 1},
	}

	for _, n := range neighbors {
		nx, ny, nz := n[0], n[1], n[2]

		var neighborBlock BlockData

		// Check if neighbor is in same chunk
		if nx >= 0 && nx < ChunkSize && ny >= 0 && ny < ChunkHeight && nz >= 0 && nz < ChunkSize {
			neighborBlock = chunk.GetBlock(nx, ny, nz)
		} else {
			// Check neighbor chunk
			worldX, worldY, worldZ := chunk.LocalToWorld(x, y, z)
			neighborBlock = cm.GetBlock(worldX+(nx-x), worldY+(ny-y), worldZ+(nz-z))
		}

		if !neighborBlock.IsSolid() {
			return true // At least one face is visible
		}
	}

	return false
}

// generateBlockMesh generates mesh data for a single block
func (cm *ChunkManager) generateBlockMesh(x, y, z int, block BlockData) ([]types.Vec3, []uint32, []types.Vec3, []types.Vec2, []types.Color) {
	// Create a hex prism for this block
	center := types.NewVec3(float32(x), float32(y), float32(z))
	prism := blocks.NewHexPrism(center, 0.5, 1.0)

	vertices := prism.GenerateVertices()
	indices := prism.GenerateIndices()
	normals := prism.GenerateNormals()
	uvs := prism.GenerateUVCoordinates()

	// Color based on block type
	color := cm.getBlockColor(block.ID)
	colors := make([]types.Color, len(vertices))
	for i := range colors {
		colors[i] = color
	}

	return vertices, indices, normals, uvs, colors
}

// getBlockColor returns the color for a block type
func (cm *ChunkManager) getBlockColor(id BlockID) types.Color {
	switch id {
	case BlockIDStone:
		return types.NewColor(128, 128, 128, 255) // Gray stone
	case BlockIDDirt:
		return types.NewColor(139, 90, 43, 255) // Brown dirt
	case BlockIDGrass:
		return types.NewColor(124, 252, 0, 255) // Green grass
	case BlockIDWood:
		return types.NewColor(160, 82, 45, 255) // Brown wood
	case BlockIDGlass:
		return types.NewColor(200, 200, 255, 180) // Light blue glass
	case BlockIDWater:
		return types.NewColor(64, 164, 223, 180) // Blue water
	default:
		return types.NewColor(255, 255, 255, 255) // White default
	}
}

// GetLoadedChunks returns all loaded chunks
func (cm *ChunkManager) GetLoadedChunks() map[ChunkCoord]*Chunk {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Return a copy
	result := make(map[ChunkCoord]*Chunk, len(cm.chunks))
	for k, v := range cm.chunks {
		result[k] = v
	}
	return result
}

// GetStats returns chunk statistics
func (cm *ChunkManager) GetStats() ChunkStats {
	cm.mu.RLock()
	cm.cacheMu.RLock()

	stats := ChunkStats{
		LoadedChunks: len(cm.chunks),
		CachedChunks: len(cm.cache),
	}

	cm.mu.RUnlock()
	cm.cacheMu.RUnlock()

	return stats
}

// SetViewDistance sets the view distance in chunks
func (cm *ChunkManager) SetViewDistance(distance int) {
	cm.mu.Lock()
	cm.viewDistance = distance
	cm.mu.Unlock()

	cm.loadChunksAroundPlayer()
	cm.unloadDistantChunks()
}

// InitializeChunkLoading triggers initial chunk loading around the given position
func (cm *ChunkManager) InitializeChunkLoading(playerPos types.Vec3) {
	// Convert player position to chunk coordinate
	playerChunkX := int(math.Floor(float64(playerPos.GetX()) / float64(ChunkSize)))
	playerChunkZ := int(math.Floor(float64(playerPos.GetZ()) / float64(ChunkSize)))

	println("📍 Initializing chunk loading around player chunk", playerChunkX, playerChunkZ)

	cm.mu.Lock()
	cm.playerChunk = ChunkCoord{X: playerChunkX, Z: playerChunkZ}
	cm.mu.Unlock()

	// Load initial chunks
	println("📦 Loading initial chunks...")
	cm.loadChunksAroundPlayer()
	println("✅ Initial chunk loading completed")
}
