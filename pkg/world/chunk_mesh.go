package world

import (
	"github.com/tesselstudio/TesselBox/pkg/blocks"
	"kaijuengine.com/matrix"
)

// ChunkMeshBuilder generates meshes for chunks
// This is a simplified implementation - in production you'd want greedy meshing

type ChunkMeshBuilder struct {
	chunk *Chunk
}

// NewChunkMeshBuilder creates a new mesh builder for a chunk
func NewChunkMeshBuilder(chunk *Chunk) *ChunkMeshBuilder {
	return &ChunkMeshBuilder{chunk: chunk}
}

// BuildMesh generates a mesh for the entire chunk
func (b *ChunkMeshBuilder) BuildMesh() *ChunkMesh {
	vertices := make([]matrix.Vec3, 0)
	indices := make([]uint32, 0)
	normals := make([]matrix.Vec3, 0)
	uvs := make([]matrix.Vec2, 0)
	colors := make([]matrix.Color, 0)

	indexOffset := uint32(0)

	// Iterate through all blocks in chunk
	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < ChunkHeight; y++ {
			for z := 0; z < ChunkSize; z++ {
				block := b.chunk.GetBlock(x, y, z)

				// Skip air blocks
				if block.IsAir() {
					continue
				}

				// Check if block is visible (has at least one air neighbor)
				if !b.isBlockVisible(x, y, z) {
					continue
				}

				// Calculate world position
				worldX := float32(b.chunk.Coord.X*ChunkSize + x)
				worldY := float32(y)
				worldZ := float32(b.chunk.Coord.Z*ChunkSize + z)

				// Get block color
				color := b.getBlockColor(block.ID)

				// Generate mesh for this block
				prism := blocks.NewHexPrism(
					matrix.NewVec3(worldX, worldY, worldZ),
					0.5, // radius
					1.0, // height
				)

				// Add prism geometry
				blockVerts := prism.GenerateVertices()
				blockIndices := prism.GenerateIndices()
				blockNormals := prism.GenerateNormals()
				blockUVs := prism.GenerateUVCoordinates()

				// Append vertices
				vertices = append(vertices, blockVerts...)

				// Append indices with offset
				for _, idx := range blockIndices {
					indices = append(indices, idx+indexOffset)
				}

				// Append normals and UVs
				normals = append(normals, blockNormals...)
				uvs = append(uvs, blockUVs...)

				// Append colors for each vertex
				for i := 0; i < len(blockVerts); i++ {
					colors = append(colors, color)
				}

				// Update index offset for next block
				indexOffset += uint32(len(blockVerts))
			}
		}
	}

	return &ChunkMesh{
		Vertices: vertices,
		Indices:  indices,
		Normals:  normals,
		UVs:      uvs,
		Colors:   colors,
	}
}

// isBlockVisible checks if a block has at least one visible face
func (b *ChunkMeshBuilder) isBlockVisible(x, y, z int) bool {
	// Check all 6 neighbors + top and bottom
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

		// Check bounds - if neighbor is outside chunk, it's visible
		if nx < 0 || nx >= ChunkSize || ny < 0 || ny >= ChunkHeight || nz < 0 || nz >= ChunkSize {
			return true
		}

		neighbor := b.chunk.GetBlock(nx, ny, nz)
		if neighbor.IsAir() || neighbor.IsTransparent() {
			return true
		}
	}

	return false
}

// getBlockColor returns the color for a block type
func (b *ChunkMeshBuilder) getBlockColor(id BlockID) matrix.Color {
	switch id {
	case BlockIDStone:
		return matrix.ColorGray()
	case BlockIDDirt:
		return matrix.NewColor(139.0/255.0, 90.0/255.0, 43.0/255.0, 1.0)
	case BlockIDGrass:
		return matrix.NewColor(124.0/255.0, 200.0/255.0, 50.0/255.0, 1.0)
	case BlockIDWood:
		return matrix.NewColor(139.0/255.0, 90.0/255.0, 43.0/255.0, 1.0)
	case BlockIDGlass:
		return matrix.NewColor(200.0/255.0, 200.0/255.0, 255.0/255.0, 0.6)
	case BlockIDWater:
		return matrix.NewColor(64.0/255.0, 164.0/255.0, 223.0/255.0, 0.8)
	default:
		return matrix.ColorWhite()
	}
}

// BuildSimpleMesh creates a simpler mesh without full hexagonal prisms
// Useful for debugging or when performance is critical
func (b *ChunkMeshBuilder) BuildSimpleMesh() *ChunkMesh {
	vertices := make([]matrix.Vec3, 0)
	indices := make([]uint32, 0)
	colors := make([]matrix.Color, 0)

	indexOffset := uint32(0)

	// Simple cube representation
	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < ChunkHeight; y++ {
			for z := 0; z < ChunkSize; z++ {
				block := b.chunk.GetBlock(x, y, z)

				if block.IsAir() {
					continue
				}

				if !b.isBlockVisible(x, y, z) {
					continue
				}

				worldX := float32(b.chunk.Coord.X*ChunkSize + x)
				worldY := float32(y)
				worldZ := float32(b.chunk.Coord.Z*ChunkSize + z)

				color := b.getBlockColor(block.ID)

				// Simple cube vertices
				blockVerts := []matrix.Vec3{
					// Front face
					matrix.NewVec3(worldX-0.5, worldY-0.5, worldZ+0.5),
					matrix.NewVec3(worldX+0.5, worldY-0.5, worldZ+0.5),
					matrix.NewVec3(worldX+0.5, worldY+0.5, worldZ+0.5),
					matrix.NewVec3(worldX-0.5, worldY+0.5, worldZ+0.5),
					// Back face
					matrix.NewVec3(worldX-0.5, worldY-0.5, worldZ-0.5),
					matrix.NewVec3(worldX-0.5, worldY+0.5, worldZ-0.5),
					matrix.NewVec3(worldX+0.5, worldY+0.5, worldZ-0.5),
					matrix.NewVec3(worldX+0.5, worldY-0.5, worldZ-0.5),
					// Top face
					matrix.NewVec3(worldX-0.5, worldY+0.5, worldZ-0.5),
					matrix.NewVec3(worldX-0.5, worldY+0.5, worldZ+0.5),
					matrix.NewVec3(worldX+0.5, worldY+0.5, worldZ+0.5),
					matrix.NewVec3(worldX+0.5, worldY+0.5, worldZ-0.5),
					// Bottom face
					matrix.NewVec3(worldX-0.5, worldY-0.5, worldZ-0.5),
					matrix.NewVec3(worldX+0.5, worldY-0.5, worldZ-0.5),
					matrix.NewVec3(worldX+0.5, worldY-0.5, worldZ+0.5),
					matrix.NewVec3(worldX-0.5, worldY-0.5, worldZ+0.5),
					// Right face
					matrix.NewVec3(worldX+0.5, worldY-0.5, worldZ-0.5),
					matrix.NewVec3(worldX+0.5, worldY+0.5, worldZ-0.5),
					matrix.NewVec3(worldX+0.5, worldY+0.5, worldZ+0.5),
					matrix.NewVec3(worldX+0.5, worldY-0.5, worldZ+0.5),
					// Left face
					matrix.NewVec3(worldX-0.5, worldY-0.5, worldZ-0.5),
					matrix.NewVec3(worldX-0.5, worldY-0.5, worldZ+0.5),
					matrix.NewVec3(worldX-0.5, worldY+0.5, worldZ+0.5),
					matrix.NewVec3(worldX-0.5, worldY+0.5, worldZ-0.5),
				}

				// Cube indices
				blockIndices := []uint32{
					// Front
					0, 1, 2, 0, 2, 3,
					// Back
					4, 5, 6, 4, 6, 7,
					// Top
					8, 9, 10, 8, 10, 11,
					// Bottom
					12, 13, 14, 12, 14, 15,
					// Right
					16, 17, 18, 16, 18, 19,
					// Left
					20, 21, 22, 20, 22, 23,
				}

				vertices = append(vertices, blockVerts...)
				for _, idx := range blockIndices {
					indices = append(indices, idx+indexOffset)
				}
				for i := 0; i < len(blockVerts); i++ {
					colors = append(colors, color)
				}

				indexOffset += uint32(len(blockVerts))
			}
		}
	}

	return &ChunkMesh{
		Vertices: vertices,
		Indices:  indices,
		Normals:  nil,
		UVs:      nil,
		Colors:   colors,
	}
}
