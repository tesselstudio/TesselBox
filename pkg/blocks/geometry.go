package blocks

import (
	"math"

	"github.com/tesselstudio/TesselBox/pkg/types"
)

// HexPrism represents a hexagonal prism with configurable properties
type HexPrism struct {
	Center   types.Vec3
	Radius   types.Float
	Height   types.Float
	Rotation int // 0-5 orientations (60-degree increments)
}

// NewHexPrism creates a new hexagonal prism with standard dimensions
func NewHexPrism(center types.Vec3, radius, height types.Float) *HexPrism {
	return &HexPrism{
		Center:   center,
		Radius:   radius,
		Height:   height,
		Rotation: 0,
	}
}

// GenerateVertices creates the vertex data for the hexagonal prism
// Returns 24 vertices: 6 top hexagon + 6 bottom hexagon + 12 side rectangles
func (h *HexPrism) GenerateVertices() []types.Vec3 {
	vertices := make([]types.Vec3, 24)

	// Generate hexagonal points
	hexPoints := h.generateHexagonPoints()

	// Top hexagon vertices (indices 0-5)
	for i := 0; i < 6; i++ {
		vertices[i] = types.NewVec3(
			hexPoints[i].X,
			h.Center.Y+float32(h.Height)/2,
			hexPoints[i].Y,
		)
	}

	// Bottom hexagon vertices (indices 6-11)
	for i := 0; i < 6; i++ {
		vertices[i+6] = types.NewVec3(
			hexPoints[i].X,
			h.Center.Y-float32(h.Height)/2,
			hexPoints[i].Y,
		)
	}

	// Side rectangle vertices (indices 12-23)
	// Each side has 4 vertices, but we'll create them as triangles later
	for i := 0; i < 6; i++ {
		next := (i + 1) % 6

		// Top edge vertices
		vertices[12+i*2] = vertices[i]      // Current top vertex
		vertices[12+i*2+1] = vertices[next] // Next top vertex
	}

	return vertices
}

// GenerateIndices creates the triangle indices for the hexagonal prism
// Returns 60 indices for 20 triangles (4 per hexagon face + 12 for sides)
func (h *HexPrism) GenerateIndices() []uint32 {
	indices := make([]uint32, 60)
	idx := 0

	// Top hexagon (4 triangles to form hexagon)
	indices[idx] = 0
	indices[idx+1] = 1
	indices[idx+2] = 2
	idx += 3

	indices[idx] = 0
	indices[idx+1] = 2
	indices[idx+2] = 3
	idx += 3

	indices[idx] = 0
	indices[idx+1] = 3
	indices[idx+2] = 4
	idx += 3

	indices[idx] = 0
	indices[idx+1] = 4
	indices[idx+2] = 5
	idx += 3

	// Bottom hexagon (4 triangles, reversed winding)
	indices[idx] = 6
	indices[idx+1] = 8
	indices[idx+2] = 7
	idx += 3

	indices[idx] = 6
	indices[idx+1] = 9
	indices[idx+2] = 8
	idx += 3

	indices[idx] = 6
	indices[idx+1] = 10
	indices[idx+2] = 9
	idx += 3

	indices[idx] = 6
	indices[idx+1] = 11
	indices[idx+2] = 10
	idx += 3

	// Side rectangles (6 rectangles, 2 triangles each)
	for i := 0; i < 6; i++ {
		next := (i + 1) % 6

		// First triangle of side
		indices[idx] = uint32(i)
		indices[idx+1] = uint32(next)
		indices[idx+2] = uint32(i + 6)
		idx += 3

		// Second triangle of side
		indices[idx] = uint32(next)
		indices[idx+1] = uint32(next + 6)
		indices[idx+2] = uint32(i + 6)
		idx += 3
	}

	return indices
}

// GenerateNormals creates normal vectors for each vertex
func (h *HexPrism) GenerateNormals() []types.Vec3 {
	normals := make([]types.Vec3, 24)

	// Top hexagon normals (pointing up)
	for i := 0; i < 6; i++ {
		normals[i] = types.NewVec3(0, 1, 0)
	}

	// Bottom hexagon normals (pointing down)
	for i := 0; i < 6; i++ {
		normals[i+6] = types.NewVec3(0, -1, 0)
	}

	// Side normals (pointing outward)
	hexPoints := h.generateHexagonPoints()
	for i := 0; i < 6; i++ {
		// Calculate outward normal for this side (already normalized)
		normal := types.NewVec3(hexPoints[i].X, 0, hexPoints[i].Y)

		// Apply to both vertices of this side
		normals[12+i*2] = normal
		normals[12+i*2+1] = normal
	}

	return normals
}

// GenerateUVCoordinates creates UV mapping for the hexagonal prism
func (h *HexPrism) GenerateUVCoordinates() []types.Vec2 {
	uvs := make([]types.Vec2, 24)

	// Top hexagon UVs
	for i := 0; i < 6; i++ {
		angle := float64(i) * math.Pi / 3.0
		u := types.Float((math.Cos(angle) + 1) * 0.5)
		v := types.Float((math.Sin(angle) + 1) * 0.5)
		uvs[i] = types.NewVec2(float32(u), float32(v))
	}

	// Bottom hexagon UVs (same as top)
	for i := 0; i < 6; i++ {
		uvs[i+6] = uvs[i]
	}

	// Side UVs (simple mapping)
	for i := 0; i < 6; i++ {
		// Each side gets a strip of UV coordinates
		uStart := types.Float(float64(i)) / 6.0
		uEnd := types.Float(float64(i+1)) / 6.0

		uvs[12+i*2] = types.NewVec2(float32(uStart), 0)
		uvs[12+i*2+1] = types.NewVec2(float32(uEnd), 0)
	}

	return uvs
}

// generateHexagonPoints creates the 2D hexagon points in the XZ plane
func (h *HexPrism) generateHexagonPoints() []types.Vec2 {
	points := make([]types.Vec2, 6)

	for i := 0; i < 6; i++ {
		angle := float64(i+h.Rotation) * math.Pi / 3.0
		x := float32(types.Float(math.Cos(angle))) * float32(h.Radius)
		z := float32(types.Float(math.Sin(angle))) * float32(h.Radius)
		points[i] = types.NewVec2(float32(x)+h.Center.X, float32(z)+h.Center.Z)
	}

	return points
}

// SetRotation sets the rotation of the hexagonal prism
func (h *HexPrism) SetRotation(rotation int) {
	h.Rotation = rotation % 6
}

// GetAttachmentFaces returns the 8 attachment faces (6 sides + 2 ends)
func (h *HexPrism) GetAttachmentFaces() []AttachmentFace {
	faces := make([]AttachmentFace, 8)

	// 6 side faces
	for i := 0; i < 6; i++ {
		faces[i] = AttachmentFace{
			Type:   FaceTypeSide,
			Index:  i,
			Normal: h.calculateSideNormal(i),
			Center: h.calculateSideCenter(i),
			Area:   h.Radius * h.Height,
		}
	}

	// Top and bottom faces
	faces[6] = AttachmentFace{
		Type:   FaceTypeTop,
		Index:  6,
		Normal: types.NewVec3(0, 1, 0),
		Center: types.NewVec3(h.Center.X, h.Center.Y+float32(h.Height)/2, h.Center.Z),
		Area:   h.calculateHexagonArea(),
	}

	faces[7] = AttachmentFace{
		Type:   FaceTypeBottom,
		Index:  7,
		Normal: types.NewVec3(0, -1, 0),
		Center: types.NewVec3(h.Center.X, h.Center.Y-float32(h.Height)/2, h.Center.Z),
		Area:   h.calculateHexagonArea(),
	}

	return faces
}

// calculateSideNormal calculates the normal vector for a side face
func (h *HexPrism) calculateSideNormal(index int) types.Vec3 {
	angle := float64(index+h.Rotation) * math.Pi / 3.0
	return types.NewVec3(
		float32(types.Float(math.Cos(angle))),
		0,
		float32(types.Float(math.Sin(angle))),
	)
}

// calculateSideCenter calculates the center point of a side face
func (h *HexPrism) calculateSideCenter(index int) types.Vec3 {
	angle := float64(index+h.Rotation) + 0.5
	angle = angle * math.Pi / 3.0
	return types.NewVec3(
		h.Center.X+float32(types.Float(math.Cos(angle)))*float32(h.Radius)*0.5,
		h.Center.Y,
		h.Center.Z+float32(types.Float(math.Sin(angle)))*float32(h.Radius)*0.5,
	)
}

// calculateHexagonArea calculates the area of the hexagonal face
func (h *HexPrism) calculateHexagonArea() types.Float {
	return types.Float(3.0*math.Sqrt(3.0)*0.5) * h.Radius * h.Radius
}

// AttachmentFace represents a face where other blocks can attach
type AttachmentFace struct {
	Type   FaceType
	Index  int
	Normal types.Vec3
	Center types.Vec3
	Area   types.Float
}

// FaceType represents the type of attachment face
type FaceType int

const (
	FaceTypeSide FaceType = iota
	FaceTypeTop
	FaceTypeBottom
)
