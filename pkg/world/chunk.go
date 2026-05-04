package world

import (
	"sync"

	"github.com/tesselstudio/TesselBox/pkg/survival"
	"kaijuengine.com/matrix"
)

// ChunkSize is the size of a chunk in blocks (16x16 hexagonal grid)
const ChunkSize = 16

// ChunkHeight is the vertical height of a chunk
const ChunkHeight = 256

// BlockID represents a block type identifier
type BlockID uint16

// Special block IDs
const (
	BlockIDAir   BlockID = 0
	BlockIDStone BlockID = 1
	BlockIDDirt  BlockID = 2
	BlockIDGrass BlockID = 3
	BlockIDWood  BlockID = 4
	BlockIDGlass BlockID = 5
	BlockIDWater BlockID = 6
)

// BlockData represents a single block in the world
type BlockData struct {
	ID       BlockID
	Metadata uint8 // Block state, rotation, etc.
	Light    uint8 // Light level 0-15
}

// IsAir returns true if this block is air
func (b BlockData) IsAir() bool {
	return b.ID == BlockIDAir
}

// IsSolid returns true if this block is solid
func (b BlockData) IsSolid() bool {
	return b.ID != BlockIDAir && b.ID != BlockIDWater
}

// IsTransparent returns true if this block is transparent
func (b BlockData) IsTransparent() bool {
	return b.ID == BlockIDAir || b.ID == BlockIDGlass || b.ID == BlockIDWater
}

// ChunkCoord represents chunk coordinates in the world
type ChunkCoord struct {
	X int // East-West chunk coordinate
	Z int // North-South chunk coordinate
}

// Chunk represents a 16x256x16 section of the world
type Chunk struct {
	mu sync.RWMutex

	Coord ChunkCoord

	// Blocks stores block data [x][y][z]
	// Using flat array for cache efficiency: blocks[x + z*ChunkSize + y*ChunkSize*ChunkSize]
	blocks []BlockData

	// Heightmap stores the highest solid block at each x,z
	heightmap [ChunkSize][ChunkSize]int16

	// Modified flag for save system
	modified bool

	// Mesh data for rendering
	meshDirty bool
	mesh      *ChunkMesh

	// Biome data
	biome survival.BiomeType
}

// ChunkMesh holds the renderable mesh data for a chunk
type ChunkMesh struct {
	Vertices []matrix.Vec3
	Indices  []uint32
	Normals  []matrix.Vec3
	UVs      []matrix.Vec2
	Colors   []matrix.Color
}

// NewChunk creates a new empty chunk
func NewChunk(coord ChunkCoord) *Chunk {
	return &Chunk{
		Coord:     coord,
		blocks:    make([]BlockData, ChunkSize*ChunkHeight*ChunkSize),
		modified:  true,
		meshDirty: true,
		biome:     survival.BiomePlains,
	}
}

// GetBlock returns the block at local coordinates
func (c *Chunk) GetBlock(x, y, z int) BlockData {
	if !c.isValidLocalCoord(x, y, z) {
		return BlockData{ID: BlockIDAir}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.blocks[c.getIndex(x, y, z)]
}

// SetBlock sets a block at local coordinates
func (c *Chunk) SetBlock(x, y, z int, block BlockData) {
	if !c.isValidLocalCoord(x, y, z) {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	idx := c.getIndex(x, y, z)
	oldBlock := c.blocks[idx]
	c.blocks[idx] = block

	if oldBlock.ID != block.ID {
		c.modified = true
		c.meshDirty = true

		// Update heightmap
		if block.IsSolid() {
			if int16(y) > c.heightmap[x][z] {
				c.heightmap[x][z] = int16(y)
			}
		} else if int16(y) == c.heightmap[x][z] {
			// Recalculate heightmap for this column
			c.recalculateHeightmapColumn(x, z)
		}
	}
}

// GetHeightmap returns the height at local x,z
func (c *Chunk) GetHeightmap(x, z int) int16 {
	if x < 0 || x >= ChunkSize || z < 0 || z >= ChunkSize {
		return 0
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.heightmap[x][z]
}

// recalculateHeightmapColumn recalculates the heightmap for a column
func (c *Chunk) recalculateHeightmapColumn(x, z int) {
	for y := ChunkHeight - 1; y >= 0; y-- {
		if c.blocks[c.getIndex(x, y, z)].IsSolid() {
			c.heightmap[x][z] = int16(y)
			return
		}
	}
	c.heightmap[x][z] = -1
}

// isValidLocalCoord checks if coordinates are within chunk bounds
func (c *Chunk) isValidLocalCoord(x, y, z int) bool {
	return x >= 0 && x < ChunkSize &&
		y >= 0 && y < ChunkHeight &&
		z >= 0 && z < ChunkSize
}

// getIndex converts local coordinates to array index
func (c *Chunk) getIndex(x, y, z int) int {
	return x + z*ChunkSize + y*ChunkSize*ChunkSize
}

// IsModified returns true if chunk has been modified since last save
func (c *Chunk) IsModified() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.modified
}

// MarkSaved marks the chunk as saved
func (c *Chunk) MarkSaved() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.modified = false
}

// IsMeshDirty returns true if chunk mesh needs regeneration
func (c *Chunk) IsMeshDirty() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.meshDirty
}

// MarkMeshClean marks the chunk mesh as clean
func (c *Chunk) MarkMeshClean() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.meshDirty = false
}

// GetMesh returns the chunk mesh
func (c *Chunk) GetMesh() *ChunkMesh {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mesh
}

// SetMesh sets the chunk mesh
func (c *Chunk) SetMesh(mesh *ChunkMesh) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mesh = mesh
	c.meshDirty = false
}

// GetBiome returns the chunk's biome
func (c *Chunk) GetBiome() survival.BiomeType {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.biome
}

// SetBiome sets the chunk's biome
func (c *Chunk) SetBiome(biome survival.BiomeType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.biome = biome
	c.modified = true
}

// GetBlockCount returns the number of non-air blocks
func (c *Chunk) GetBlockCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, block := range c.blocks {
		if block.ID != BlockIDAir {
			count++
		}
	}
	return count
}

// Fill fills the chunk with a specific block
func (c *Chunk) Fill(block BlockData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.blocks {
		c.blocks[i] = block
	}

	if block.IsSolid() {
		for x := 0; x < ChunkSize; x++ {
			for z := 0; z < ChunkSize; z++ {
				c.heightmap[x][z] = ChunkHeight - 1
			}
		}
	} else {
		for x := 0; x < ChunkSize; x++ {
			for z := 0; z < ChunkSize; z++ {
				c.heightmap[x][z] = -1
			}
		}
	}

	c.modified = true
	c.meshDirty = true
}

// LocalToWorld converts local chunk coordinates to world coordinates
func (c *Chunk) LocalToWorld(x, y, z int) (worldX, worldY, worldZ int) {
	worldX = c.Coord.X*ChunkSize + x
	worldY = y
	worldZ = c.Coord.Z*ChunkSize + z
	return
}

// WorldToLocal converts world coordinates to local chunk coordinates
func WorldToLocal(worldX, worldY, worldZ int) (chunkCoord ChunkCoord, localX, localY, localZ int) {
	chunkCoord.X = worldX / ChunkSize
	chunkCoord.Z = worldZ / ChunkSize

	localX = worldX % ChunkSize
	localY = worldY
	localZ = worldZ % ChunkSize

	if worldX < 0 {
		chunkCoord.X--
		localX += ChunkSize
	}
	if worldZ < 0 {
		chunkCoord.Z--
		localZ += ChunkSize
	}

	return
}
