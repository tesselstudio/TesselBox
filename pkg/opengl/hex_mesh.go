package opengl

import (
	"math"

	"github.com/tesselstudio/TesselBox/pkg/types"
	"github.com/tesselstudio/TesselBox/pkg/world"
)

// ChunkMeshData holds vertex and index data for a chunk
type ChunkMeshData struct {
	Vertices    []float32
	Indices     []uint32
	VertexCount int32
	IndexCount  int32
}

// HexMeshGenerator generates mesh data for hexagonal chunks
type HexMeshGenerator struct {
	chunk *world.Chunk
}

// NewHexMeshGenerator creates a new mesh generator
func NewHexMeshGenerator(chunk *world.Chunk) *HexMeshGenerator {
	return &HexMeshGenerator{
		chunk: chunk,
	}
}

// GenerateMesh generates mesh data for the chunk
func (hmg *HexMeshGenerator) GenerateMesh() *ChunkMeshData {
	meshData := &ChunkMeshData{
		Vertices: make([]float32, 0, 4096),
		Indices:  make([]uint32, 0, 2048),
	}

	// Iterate through all blocks in the chunk
	for x := 0; x < world.ChunkSize; x++ {
		for y := 0; y < world.ChunkHeight; y++ {
			for z := 0; z < world.ChunkSize; z++ {
				block := hmg.chunk.GetBlock(x, y, z)
				if block.ID == 0 { // Air block
					continue
				}

				// Generate hexagonal prism for this block
				blockPos := types.NewVec3(float32(x), float32(y), float32(z))
				hmg.addHexPrismFaces(meshData, blockPos, block.ID)
			}
		}
	}

	meshData.VertexCount = int32(len(meshData.Vertices) / 6) // 6 values per vertex (pos + color)
	meshData.IndexCount = int32(len(meshData.Indices))

	return meshData
}

// addHexPrismFaces adds the faces of a hexagonal prism to the mesh
func (hmg *HexMeshGenerator) addHexPrismFaces(meshData *ChunkMeshData, blockPos types.Vec3, blockID world.BlockID) {
	// Hexagonal prism vertices (normalized radius 1.0, height 2.0)
	radius := 0.5
	height := 1.0

	// Calculate hexagon vertices (6 vertices around)
	hexVertices := make([]types.Vec3, 6)
	for i := 0; i < 6; i++ {
		angle := float64(i) * math.Pi / 3.0 // 60 degrees between vertices
		x := float64(blockPos.X) + radius*math.Cos(angle)
		z := float64(blockPos.Z) + radius*math.Sin(angle)
		hexVertices[i] = types.NewVec3(float32(x), blockPos.Y, float32(z))
	}

	// Get block color based on ID
	color := hmg.getBlockColor(blockID)

	// Add top face (hexagon)
	topBaseIndex := uint32(len(meshData.Vertices) / 6)
	for i := 0; i < 6; i++ {
		hmg.addVertex(meshData, hexVertices[i], blockPos.Y+float32(height), color)
	}
	for i := 1; i < 5; i++ {
		meshData.Indices = append(meshData.Indices, topBaseIndex, topBaseIndex+uint32(i), topBaseIndex+uint32(i+1))
	}

	// Add bottom face (hexagon)
	bottomBaseIndex := uint32(len(meshData.Vertices) / 6)
	for i := 0; i < 6; i++ {
		hmg.addVertex(meshData, hexVertices[i], blockPos.Y-float32(height), color)
	}
	for i := 1; i < 5; i++ {
		meshData.Indices = append(meshData.Indices, bottomBaseIndex, bottomBaseIndex+uint32(i+1), bottomBaseIndex+uint32(i))
	}

	// Add side faces (6 rectangular faces)
	for i := 0; i < 6; i++ {
		nextI := (i + 1) % 6

		// Get base indices for top and bottom vertices
		topIdx := topBaseIndex + uint32(i)
		topNextIdx := topBaseIndex + uint32(nextI)
		bottomIdx := bottomBaseIndex + uint32(i)
		bottomNextIdx := bottomBaseIndex + uint32(nextI)

		// Add quad as two triangles
		meshData.Indices = append(meshData.Indices, topIdx, bottomIdx, bottomNextIdx)
		meshData.Indices = append(meshData.Indices, topIdx, bottomNextIdx, topNextIdx)
	}
}

// addVertex adds a vertex to the mesh
func (hmg *HexMeshGenerator) addVertex(meshData *ChunkMeshData, pos types.Vec3, y float32, color types.Vec3) {
	// Position (3 floats)
	meshData.Vertices = append(meshData.Vertices, pos.X, y, pos.Z)
	// Color (3 floats)
	meshData.Vertices = append(meshData.Vertices, color.X, color.Y, color.Z)
}

// getBlockColor returns the color for a block type
func (hmg *HexMeshGenerator) getBlockColor(blockID world.BlockID) types.Vec3 {
	switch blockID {
	case 1: // Stone
		return types.NewVec3(0.8, 0.8, 0.8)
	case 2: // Dirt
		return types.NewVec3(0.6, 0.45, 0.25)
	case 3: // Grass
		return types.NewVec3(0.2, 0.7, 0.2)
	case 4: // Wood
		return types.NewVec3(0.6, 0.35, 0.15)
	case 5: // Leaves
		return types.NewVec3(0.2, 0.6, 0.2)
	case 6: // Sand
		return types.NewVec3(0.95, 0.95, 0.7)
	case 7: // Water
		return types.NewVec3(0.3, 0.5, 1.0)
	case 8: // Lava
		return types.NewVec3(1.0, 0.5, 0.0)
	default:
		return types.NewVec3(1.0, 1.0, 1.0) // Default white
	}
}
