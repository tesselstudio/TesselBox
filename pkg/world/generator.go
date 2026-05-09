package world

import (
	"math"
	"sync"
)

// WorldGenerator generates terrain for the world using noise
type WorldGenerator struct {
	mu sync.Mutex

	seed     int64
	seaLevel int

	// Noise generators for different terrain features
	heightNoise  *SimplexNoise
	biomeNoise   *SimplexNoise
	caveNoise    *SimplexNoise
	featureNoise *SimplexNoise
}

// NewWorldGenerator creates a new world generator with the given seed
func NewWorldGenerator(seed int64) *WorldGenerator {
	return &WorldGenerator{
		seed:         seed,
		seaLevel:     64,
		heightNoise:  NewSimplexNoise(seed),
		biomeNoise:   NewSimplexNoise(seed + 1),
		caveNoise:    NewSimplexNoise(seed + 2),
		featureNoise: NewSimplexNoise(seed + 3),
	}
}

// GenerateChunk generates terrain for a chunk
func (wg *WorldGenerator) GenerateChunk(chunk *Chunk) {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	baseX := chunk.Coord.X * ChunkSize
	baseZ := chunk.Coord.Z * ChunkSize

	// Generate biome for this chunk
	biome := wg.determineBiome(baseX+ChunkSize/2, baseZ+ChunkSize/2)
	chunk.SetBiome(biome)

	// Generate terrain heightmap
	heightMap := wg.generateHeightmap(baseX, baseZ, biome)

	// Fill blocks based on heightmap
	for x := 0; x < ChunkSize; x++ {
		for z := 0; z < ChunkSize; z++ {
			height := heightMap[x][z]

			// Fill column
			for y := 0; y < ChunkHeight && y <= height+3; y++ {
				block := wg.determineBlock(x+baseX, y, z+baseZ, height, biome)
				chunk.SetBlock(x, y, z, block)
			}
		}
	}

	// Generate caves and features
	wg.generateCaves(chunk, baseX, baseZ)
	wg.generateFeatures(chunk, baseX, baseZ, biome)
}

// generateHeightmap generates a heightmap for the chunk
func (wg *WorldGenerator) generateHeightmap(baseX, baseZ int, biome BiomeType) [ChunkSize][ChunkSize]int {
	var heightMap [ChunkSize][ChunkSize]int

	for x := 0; x < ChunkSize; x++ {
		for z := 0; z < ChunkSize; z++ {
			worldX := float64(baseX + x)
			worldZ := float64(baseZ + z)

			// Base terrain noise
			height := wg.heightNoise.Noise2D(worldX*0.01, worldZ*0.01)
			height += wg.heightNoise.Noise2D(worldX*0.05, worldZ*0.05) * 0.5
			height += wg.heightNoise.Noise2D(worldX*0.1, worldZ*0.1) * 0.25

			// Normalize to 0-1
			height = (height + 1.0) * 0.5

			// Apply biome height modifiers
			baseHeight, heightVar := wg.getBiomeHeightRange(biome)
			finalHeight := baseHeight + int(height*float64(heightVar))

			heightMap[x][z] = finalHeight
		}
	}

	return heightMap
}

// determineBlock determines what block type to place at a position
func (wg *WorldGenerator) determineBlock(x, y, z int, groundHeight int, biome BiomeType) BlockData {
	// Water below sea level
	if y <= wg.seaLevel && y > groundHeight {
		return BlockData{ID: BlockIDWater}
	}

	// Above ground = air
	if y > groundHeight {
		return BlockData{ID: BlockIDAir}
	}

	// Surface blocks based on depth
	depth := groundHeight - y

	if depth == 0 {
		// Surface block based on biome
		switch biome {
		case BiomeDesert:
			return BlockData{ID: BlockIDDirt} // Sand would be a new block type
		case BiomeTundra:
			return BlockData{ID: BlockIDDirt} // Snow covered
		default:
			return BlockData{ID: BlockIDGrass}
		}
	} else if depth <= 3 {
		return BlockData{ID: BlockIDDirt}
	} else {
		return BlockData{ID: BlockIDStone}
	}
}

// determineBiome determines the biome at world coordinates
func (wg *WorldGenerator) determineBiome(x, z int) BiomeType {
	worldX := float64(x)
	worldZ := float64(z)

	// Temperature and humidity noise
	temp := wg.biomeNoise.Noise2D(worldX*0.002, worldZ*0.002)
	humidity := wg.biomeNoise.Noise2D(worldX*0.002+1000, worldZ*0.002+1000)

	// Determine biome based on temperature and humidity
	switch {
	case temp > 0.5:
		// Hot
		if humidity < -0.2 {
			return BiomeDesert
		}
		return BiomeJungle

	case temp < -0.3:
		// Cold
		if humidity < 0 {
			return BiomeTundra
		}
		return BiomeMountains

	default:
		// Temperate
		if humidity > 0.3 {
			return BiomeForest
		} else if humidity < -0.3 {
			return BiomePlains
		}
		return BiomeForest
	}
}

// getBiomeHeightRange returns base height and variation for a biome
func (wg *WorldGenerator) getBiomeHeightRange(biome BiomeType) (baseHeight int, heightVar int) {
	switch biome {
	case BiomePlains:
		return 64, 10
	case BiomeForest:
		return 68, 15
	case BiomeDesert:
		return 62, 8
	case BiomeTundra:
		return 70, 20
	case BiomeMountains:
		return 80, 60
	case BiomeJungle:
		return 65, 12
	default:
		return 64, 10
	}
}

// generateCaves generates caves in the chunk
func (wg *WorldGenerator) generateCaves(chunk *Chunk, baseX, baseZ int) {
	for x := 0; x < ChunkSize; x++ {
		for z := 0; z < ChunkSize; z++ {
			for y := 5; y < 60; y++ { // Caves only below surface
				worldX := float64(baseX + x)
				worldY := float64(y)
				worldZ := float64(baseZ + z)

				// 3D noise for caves
				caveValue := wg.caveNoise.Noise3D(worldX*0.05, worldY*0.05, worldZ*0.05)

				// Create cave if noise value is high enough
				if caveValue > 0.7 {
					chunk.SetBlock(x, y, z, BlockData{ID: BlockIDAir})
				}
			}
		}
	}
}

// generateFeatures generates trees, ores, and other features
func (wg *WorldGenerator) generateFeatures(chunk *Chunk, baseX, baseZ int, biome BiomeType) {
	// Generate ores
	wg.generateOres(chunk, baseX, baseZ)

	// Generate trees (simplified - just place wood blocks)
	wg.generateTrees(chunk, baseX, baseZ, biome)
}

// generateOres generates ore deposits
func (wg *WorldGenerator) generateOres(chunk *Chunk, baseX, baseZ int) {
	// Coal ore - common, all depths
	wg.generateOreVein(chunk, baseX, baseZ, 15, 80, BlockIDStone, 0.3, 8)

	// Iron ore - less common, mid depths
	wg.generateOreVein(chunk, baseX, baseZ, 10, 60, BlockIDStone, 0.2, 6)

	// Gold ore - rare, deep
	wg.generateOreVein(chunk, baseX, baseZ, 5, 30, BlockIDStone, 0.1, 4)
}

// generateOreVein generates a vein of ore
func (wg *WorldGenerator) generateOreVein(chunk *Chunk, baseX, baseZ, minY, maxY int, blockType BlockID, chance float64, veinSize int) {
	for x := 0; x < ChunkSize; x++ {
		for z := 0; z < ChunkSize; z++ {
			worldX := float64(baseX + x)
			worldZ := float64(baseZ + z)

			// Check if this chunk contains ore
			oreValue := wg.featureNoise.Noise2D(worldX*0.1, worldZ*0.1)

			if oreValue > 1.0-chance {
				// Generate ore vein starting here
				startY := minY + int(wg.featureNoise.Noise2D(worldX, worldZ)*(float64(maxY-minY)))

				for i := 0; i < veinSize && startY+i < ChunkHeight; i++ {
					y := startY + i
					block := chunk.GetBlock(x, y, z)
					if block.ID == BlockIDStone {
						// Replace stone with "ore" - using colored stone for now
						chunk.SetBlock(x, y, z, BlockData{ID: blockType, Metadata: uint8(blockType)})
					}
				}
			}
		}
	}
}

// generateTrees generates trees based on biome
func (wg *WorldGenerator) generateTrees(chunk *Chunk, baseX, baseZ int, biome BiomeType) {
	// Tree density based on biome
	treeChance := 0.0
	switch biome {
	case BiomeForest:
		treeChance = 0.1
	case BiomeJungle:
		treeChance = 0.15
	case BiomePlains:
		treeChance = 0.02
	default:
		treeChance = 0.0
	}

	if treeChance == 0.0 {
		return
	}

	for x := 2; x < ChunkSize-2; x++ {
		for z := 2; z < ChunkSize-2; z++ {
			worldX := float64(baseX + x)
			worldZ := float64(baseZ + z)

			treeValue := wg.featureNoise.Noise2D(worldX*0.5, worldZ*0.5)

			if treeValue > 1.0-treeChance {
				// Find ground height
				groundY := int(chunk.GetHeightmap(x, z))

				// Place tree trunk
				treeHeight := 4 + int(wg.featureNoise.Noise2D(worldX, worldZ)*3)
				for h := 0; h < treeHeight && groundY+h < ChunkHeight; h++ {
					chunk.SetBlock(x, groundY+1+h, z, BlockData{ID: BlockIDWood})
				}
			}
		}
	}
}

// SimplexNoise is a simple noise generator (simplified implementation)
type SimplexNoise struct {
	seed int64
	perm []int
}

// NewSimplexNoise creates a new simplex noise generator
func NewSimplexNoise(seed int64) *SimplexNoise {
	perm := make([]int, 512)
	for i := 0; i < 256; i++ {
		perm[i] = i
	}

	// Shuffle based on seed
	s := seed
	for i := 255; i > 0; i-- {
		s = (s*1103515245 + 12345) & 0x7fffffff
		j := int(s) % (i + 1)
		perm[i], perm[j] = perm[j], perm[i]
	}

	// Duplicate for speed
	for i := 0; i < 256; i++ {
		perm[i+256] = perm[i]
	}

	return &SimplexNoise{seed: seed, perm: perm}
}

// Noise2D generates 2D noise
func (sn *SimplexNoise) Noise2D(x, y float64) float64 {
	// Simplified value noise for now
	// In production, replace with proper simplex noise

	ix := int(math.Floor(x))
	iy := int(math.Floor(y))

	fx := x - float64(ix)
	fy := y - float64(iy)

	// Hash function
	hash := func(x, y int) int {
		return sn.perm[(x&255)+sn.perm[(y&255)&255]]
	}

	// Get corner values
	n00 := float64(hash(ix, iy))/255.0*2.0 - 1.0
	n10 := float64(hash(ix+1, iy))/255.0*2.0 - 1.0
	n01 := float64(hash(ix, iy+1))/255.0*2.0 - 1.0
	n11 := float64(hash(ix+1, iy+1))/255.0*2.0 - 1.0

	// Bilinear interpolation
	n0 := n00*(1.0-fx) + n10*fx
	n1 := n01*(1.0-fx) + n11*fx

	return n0*(1.0-fy) + n1*fy
}

// Noise3D generates 3D noise
func (sn *SimplexNoise) Noise3D(x, y, z float64) float64 {
	// Simplified 3D noise - average of 2D noise at different offsets
	n1 := sn.Noise2D(x, y)
	n2 := sn.Noise2D(y, z)
	n3 := sn.Noise2D(x, z)
	return (n1 + n2 + n3) / 3.0
}
