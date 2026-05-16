package world

import (
	"math"

	"github.com/tesselstudio/TesselBox/pkg/blocks"
	"github.com/tesselstudio/TesselBox/pkg/types"
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
	vertices := make([]types.Vec3, 0)
	indices := make([]uint32, 0)
	normals := make([]types.Vec3, 0)
	uvs := make([]types.Vec2, 0)
	colors := make([]types.Color, 0)

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

				// Generate mesh for this block
				prism := blocks.NewHexPrism(
					types.NewVec3(worldX, worldY, worldZ),
					0.5, // radius
					1.0, // height
				)

				// Generate face-based colors for better 3D visualization
				faceColors := b.generateFaceColors(block.ID)

				// Save current vertex count as base index
				baseIndex := uint32(len(vertices))

				// Append vertices
				blockVerts := prism.GenerateVertices()
				vertices = append(vertices, blockVerts...)

				// Append indices
				blockIndices := prism.GenerateIndices()
				for _, index := range blockIndices {
					indices = append(indices, baseIndex+index)
				}

				// Append normals
				blockNormals := prism.GenerateNormals()
				normals = append(normals, blockNormals...)

				// Append face-specific colors
				colors = append(colors, faceColors...)

				// Update index offset for next block
				indexOffset += uint32(len(blockVerts))
			}
		}
	}

	// Convert to interleaved vertex format expected by OpenGL
	interleavedVertices := make([]float32, 0, len(vertices)*6)
	for i := 0; i < len(vertices); i++ {
		// Position (3 floats)
		interleavedVertices = append(interleavedVertices, vertices[i].X, vertices[i].Y, vertices[i].Z)
		// Color (3 floats) - convert from Color type
		if i < len(colors) {
			interleavedVertices = append(interleavedVertices,
				float32(colors[i].R)/255.0,
				float32(colors[i].G)/255.0,
				float32(colors[i].B)/255.0)
		} else {
			interleavedVertices = append(interleavedVertices, 1.0, 1.0, 1.0)
		}
	}

	return &ChunkMesh{
		Vertices:            vertices,
		Indices:             indices,
		Normals:             normals,
		UVs:                 uvs,
		Colors:              colors,
		InterleavedVertices: interleavedVertices,
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

// generateFaceColors creates different colors for each face of a block
func (b *ChunkMeshBuilder) generateFaceColors(id BlockID) []types.Color {
	colors := make([]types.Color, 12) // 12 vertices for hexagonal prism (6 top + 6 bottom)

	// Base color for this block type
	baseColor := b.getBlockColor(id)

	// Top face (vertices 0-5) - lighter
	topColor := lightenColor(baseColor, 0.3)
	for i := 0; i < 6; i++ {
		colors[i] = topColor
	}

	// Bottom face (vertices 6-11) - darker
	bottomColor := darkenColor(baseColor, 0.3)
	for i := 6; i < 12; i++ {
		colors[i] = bottomColor
	}

	return colors
}

// lightenColor makes a color lighter
func lightenColor(color types.Color, factor float32) types.Color {
	return types.NewColor(
		uint8(float32(color.R)+(255-float32(color.R))*factor),
		uint8(float32(color.G)+(255-float32(color.G))*factor),
		uint8(float32(color.B)+(255-float32(color.B))*factor),
		color.A,
	)
}

// darkenColor makes a color darker
func darkenColor(color types.Color, factor float32) types.Color {
	return types.NewColor(
		uint8(float32(color.R)*(1-factor)),
		uint8(float32(color.G)*(1-factor)),
		uint8(float32(color.B)*(1-factor)),
		color.A,
	)
}

// adjustHue adjusts the hue of a color for variety
func adjustHue(baseColor types.Color, hueShift float32) types.Color {
	// Simple hue adjustment by shifting RGB values
	r := float32(baseColor.R)
	g := float32(baseColor.G)
	b := float32(baseColor.B)

	// Apply hue shift (simplified)
	shiftR := float32(math.Cos(float64(hueShift))) * 0.5
	shiftG := float32(math.Cos(float64(hueShift)-2.094)) * 0.5
	shiftB := float32(math.Cos(float64(hueShift)+2.094)) * 0.5

	newR := r + (shiftR-0.5)*60
	newG := g + (shiftG-0.5)*60
	newB := b + (shiftB-0.5)*60

	// Clamp values
	if newR > 255 {
		newR = 255
	}
	if newG > 255 {
		newG = 255
	}
	if newB > 255 {
		newB = 255
	}
	if newR < 0 {
		newR = 0
	}
	if newG < 0 {
		newG = 0
	}
	if newB < 0 {
		newB = 0
	}

	return types.NewColor(uint8(newR), uint8(newG), uint8(newB), baseColor.A)
}

// getBlockColor returns the color for a block type
func (b *ChunkMeshBuilder) getBlockColor(id BlockID) types.Color {
	// Use different colors for different block types for better visualization
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

// BuildSimpleMesh creates a simpler mesh without full hexagonal prisms
// Useful for debugging or when performance is critical
func (b *ChunkMeshBuilder) BuildSimpleMesh() *ChunkMesh {
	vertices := make([]types.Vec3, 0)
	indices := make([]uint32, 0)
	colors := make([]types.Color, 0)

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
				blockVerts := []types.Vec3{
					// Front face
					types.NewVec3(worldX-0.5, worldY-0.5, worldZ+0.5),
					types.NewVec3(worldX+0.5, worldY-0.5, worldZ+0.5),
					types.NewVec3(worldX+0.5, worldY+0.5, worldZ+0.5),
					types.NewVec3(worldX-0.5, worldY+0.5, worldZ+0.5),
					// Back face
					types.NewVec3(worldX-0.5, worldY-0.5, worldZ-0.5),
					types.NewVec3(worldX-0.5, worldY+0.5, worldZ-0.5),
					types.NewVec3(worldX+0.5, worldY+0.5, worldZ-0.5),
					types.NewVec3(worldX+0.5, worldY-0.5, worldZ-0.5),
					// Top face
					types.NewVec3(worldX-0.5, worldY+0.5, worldZ-0.5),
					types.NewVec3(worldX-0.5, worldY+0.5, worldZ+0.5),
					types.NewVec3(worldX+0.5, worldY+0.5, worldZ+0.5),
					types.NewVec3(worldX+0.5, worldY+0.5, worldZ-0.5),
					// Bottom face
					types.NewVec3(worldX-0.5, worldY-0.5, worldZ-0.5),
					types.NewVec3(worldX+0.5, worldY-0.5, worldZ-0.5),
					types.NewVec3(worldX+0.5, worldY-0.5, worldZ+0.5),
					types.NewVec3(worldX-0.5, worldY-0.5, worldZ+0.5),
					// Right face
					types.NewVec3(worldX+0.5, worldY-0.5, worldZ-0.5),
					types.NewVec3(worldX+0.5, worldY+0.5, worldZ-0.5),
					types.NewVec3(worldX+0.5, worldY+0.5, worldZ+0.5),
					types.NewVec3(worldX+0.5, worldY-0.5, worldZ+0.5),
					// Left face
					types.NewVec3(worldX-0.5, worldY-0.5, worldZ-0.5),
					types.NewVec3(worldX-0.5, worldY-0.5, worldZ+0.5),
					types.NewVec3(worldX-0.5, worldY+0.5, worldZ+0.5),
					types.NewVec3(worldX-0.5, worldY+0.5, worldZ-0.5),
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
