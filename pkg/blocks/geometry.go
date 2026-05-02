package blocks

import (
	"math"

	"kaijuengine.com/matrix"
)

// HexPrism represents a hexagonal prism with configurable properties
type HexPrism struct {
	Center   matrix.Vec3
	Radius   matrix.Float
	Height   matrix.Float
	Rotation int // 0-5 orientations (60-degree increments)
}

// NewHexPrism creates a new hexagonal prism with standard dimensions
func NewHexPrism(center matrix.Vec3, radius, height matrix.Float) *HexPrism {
	return &HexPrism{
		Center:   center,
		Radius:   radius,
		Height:   height,
		Rotation: 0,
	}
}

// GenerateVertices creates the vertex data for the hexagonal prism
// Returns 24 vertices: 6 top hexagon + 6 bottom hexagon + 12 side rectangles
func (h *HexPrism) GenerateVertices() []matrix.Vec3 {
	vertices := make([]matrix.Vec3, 24)

	// Generate hexagonal points
	hexPoints := h.generateHexagonPoints()

	// Top hexagon vertices (indices 0-5)
	for i := 0; i < 6; i++ {
		vertices[i] = matrix.NewVec3(
			hexPoints[i][0],
			h.Center.Y()+h.Height/2,
			hexPoints[i][1],
		)
	}

	// Bottom hexagon vertices (indices 6-11)
	for i := 0; i < 6; i++ {
		vertices[i+6] = matrix.NewVec3(
			hexPoints[i][0],
			h.Center.Y()-h.Height/2,
			hexPoints[i][1],
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
// Returns 36 indices for 12 triangles (2 per hexagon face + 6 for sides)
func (h *HexPrism) GenerateIndices() []uint32 {
	indices := make([]uint32, 36)
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
func (h *HexPrism) GenerateNormals() []matrix.Vec3 {
	normals := make([]matrix.Vec3, 24)

	// Top hexagon normals (pointing up)
	for i := 0; i < 6; i++ {
		normals[i] = matrix.NewVec3(0, 1, 0)
	}

	// Bottom hexagon normals (pointing down)
	for i := 0; i < 6; i++ {
		normals[i+6] = matrix.NewVec3(0, -1, 0)
	}

	// Side normals (pointing outward)
	hexPoints := h.generateHexagonPoints()
	for i := 0; i < 6; i++ {
		// Calculate outward normal for this side
		normal := matrix.NewVec3(hexPoints[i][0], 0, hexPoints[i][1])
		normal = normal.Normal()

		// Apply to both vertices of this side
		normals[12+i*2] = normal
		normals[12+i*2+1] = normal
	}

	return normals
}

// GenerateUVCoordinates creates UV mapping for the hexagonal prism
func (h *HexPrism) GenerateUVCoordinates() []matrix.Vec2 {
	uvs := make([]matrix.Vec2, 24)

	// Top hexagon UVs
	for i := 0; i < 6; i++ {
		angle := float64(i) * math.Pi / 3.0
		u := matrix.Float((math.Cos(angle) + 1) * 0.5)
		v := matrix.Float((math.Sin(angle) + 1) * 0.5)
		uvs[i] = matrix.NewVec2(u, v)
	}

	// Bottom hexagon UVs (same as top)
	for i := 0; i < 6; i++ {
		uvs[i+6] = uvs[i]
	}

	// Side UVs (simple mapping)
	for i := 0; i < 6; i++ {
		// Each side gets a strip of UV coordinates
		uStart := matrix.Float(float64(i)) / 6.0
		uEnd := matrix.Float(float64(i+1)) / 6.0

		uvs[12+i*2] = matrix.NewVec2(uStart, 0)
		uvs[12+i*2+1] = matrix.NewVec2(uEnd, 0)
	}

	return uvs
}

// generateHexagonPoints creates the 2D hexagon points in the XZ plane
func (h *HexPrism) generateHexagonPoints() []matrix.Vec2 {
	points := make([]matrix.Vec2, 6)

	for i := 0; i < 6; i++ {
		angle := float64(i+h.Rotation) * math.Pi / 3.0
		x := matrix.Float(math.Cos(angle)) * h.Radius
		z := matrix.Float(math.Sin(angle)) * h.Radius
		points[i] = matrix.NewVec2(x+h.Center.X(), z+h.Center.Z())
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
		Normal: matrix.NewVec3(0, 1, 0),
		Center: matrix.NewVec3(h.Center.X(), h.Center.Y()+h.Height/2, h.Center.Z()),
		Area:   h.calculateHexagonArea(),
	}

	faces[7] = AttachmentFace{
		Type:   FaceTypeBottom,
		Index:  7,
		Normal: matrix.NewVec3(0, -1, 0),
		Center: matrix.NewVec3(h.Center.X(), h.Center.Y()-h.Height/2, h.Center.Z()),
		Area:   h.calculateHexagonArea(),
	}

	return faces
}

// calculateSideNormal calculates the normal vector for a side face
func (h *HexPrism) calculateSideNormal(index int) matrix.Vec3 {
	angle := float64(index+h.Rotation) * math.Pi / 3.0
	return matrix.NewVec3(
		matrix.Float(math.Cos(angle)),
		0,
		matrix.Float(math.Sin(angle)),
	)
}

// calculateSideCenter calculates the center point of a side face
func (h *HexPrism) calculateSideCenter(index int) matrix.Vec3 {
	angle := float64(index+h.Rotation) + 0.5
	angle = angle * math.Pi / 3.0
	return matrix.NewVec3(
		h.Center.X()+matrix.Float(math.Cos(angle))*h.Radius*0.5,
		h.Center.Y(),
		h.Center.Z()+matrix.Float(math.Sin(angle))*h.Radius*0.5,
	)
}

// calculateHexagonArea calculates the area of the hexagonal face
func (h *HexPrism) calculateHexagonArea() matrix.Float {
	return matrix.Float(3.0*math.Sqrt(3.0)*0.5) * h.Radius * h.Radius
}

// AttachmentFace represents a face where other blocks can attach
type AttachmentFace struct {
	Type   FaceType
	Index  int
	Normal matrix.Vec3
	Center matrix.Vec3
	Area   matrix.Float
}

// FaceType represents the type of attachment face
type FaceType int

const (
	FaceTypeSide FaceType = iota
	FaceTypeTop
	FaceTypeBottom
)
