// Package world provides hexagonal coordinate system and world management
// for the TesselBox voxel game.
//
// The coordinate system uses axial coordinates (Q, R) for hexagonal grids,
// providing efficient neighbor calculations and distance measurements.
//
// # Coordinate System
//
// HexCoord represents a position in hexagonal axial coordinates:
//   - Q: Column coordinate (x-axis)
//   - R: Row coordinate (diagonal axis)
//
// Usage
//
//	coord := world.HexCoord{Q: 5, R: 3}
//	worldPos := coord.ToWorld(1.0) // Convert to world space
//	distance := coord.Distance(other) // Calculate hex distance
//	neighbor := coord.Neighbor(world.HexDirectionEast) // Get neighbor
//
// # Thread Safety
//
// All coordinate operations are thread-safe as they work with immutable values.
package world

import (
	"math"
	"strconv"

	"kaijuengine.com/matrix"
)

// HexCoord represents a position in hexagonal axial coordinates
type HexCoord struct {
	Q int // Column
	R int // Row
}

// NewHexCoord creates a new hexagonal coordinate
func NewHexCoord(q, r int) HexCoord {
	return HexCoord{Q: q, R: r}
}

// ToWorld converts hexagonal coordinates to 3D world coordinates
func (h HexCoord) ToWorld(hexSize matrix.Float) matrix.Vec3 {
	x := matrix.Float(h.Q) * hexSize * 1.5
	z := matrix.Float(h.Q+h.R) * hexSize * matrix.Float(math.Sqrt(3)*0.5)

	return matrix.NewVec3(x, 0, z)
}

// ToWorldWithHeight converts hexagonal coordinates to 3D world coordinates with height
func (h HexCoord) ToWorldWithHeight(hexSize, height matrix.Float) matrix.Vec3 {
	world := h.ToWorld(hexSize)
	world.SetY(height)
	return world
}

// Distance calculates the distance between two hexagonal coordinates
func (h HexCoord) Distance(other HexCoord) int {
	return (abs(h.Q-other.Q) + abs(h.Q+h.R-other.Q-other.R) + abs(h.R-other.R)) / 2
}

// Neighbor returns the neighboring hexagonal coordinate in the specified direction
func (h HexCoord) Neighbor(direction HexDirection) HexCoord {
	return HexCoord{
		Q: h.Q + direction.DQ(),
		R: h.R + direction.DR(),
	}
}

// GetNeighbors returns all 6 neighboring coordinates
func (h HexCoord) GetNeighbors() []HexCoord {
	neighbors := make([]HexCoord, 6)
	for i, dir := range HexDirections {
		neighbors[i] = h.Neighbor(dir)
	}
	return neighbors
}

// Ring returns all coordinates at the specified distance from this coordinate
func (h HexCoord) Ring(radius int) []HexCoord {
	if radius == 0 {
		return []HexCoord{h}
	}

	results := make([]HexCoord, 0, radius*6)

	// Start from one of the directions
	current := h.Neighbor(HexDirectionEast).RingStep(radius-1, HexDirectionSouthEast)

	// Walk around the ring
	for _, dir := range HexDirections {
		for i := 0; i < radius; i++ {
			results = append(results, current)
			current = current.Neighbor(dir)
		}
	}

	return results
}

// RingStep moves multiple steps in a direction, used for ring generation
func (h HexCoord) RingStep(steps int, direction HexDirection) HexCoord {
	result := h
	for i := 0; i < steps; i++ {
		result = result.Neighbor(direction)
	}
	return result
}

// Spiral returns coordinates in a spiral pattern around this coordinate
func (h HexCoord) Spiral(maxRadius int) []HexCoord {
	results := make([]HexCoord, 0, 1+3*maxRadius*(maxRadius+1))

	for radius := 0; radius <= maxRadius; radius++ {
		ringCoords := h.Ring(radius)
		results = append(results, ringCoords...)
	}

	return results
}

// Rotate rotates this coordinate around the origin by 60-degree increments
func (h HexCoord) Rotate(rotations int) HexCoord {
	rotations = rotations % 6

	for i := 0; i < rotations; i++ {
		h = HexCoord{
			Q: -h.R,
			R: h.Q + h.R,
		}
	}

	return h
}

// Add adds another hexagonal coordinate to this one
func (h HexCoord) Add(other HexCoord) HexCoord {
	return HexCoord{
		Q: h.Q + other.Q,
		R: h.R + other.R,
	}
}

// Subtract subtracts another hexagonal coordinate from this one
func (h HexCoord) Subtract(other HexCoord) HexCoord {
	return HexCoord{
		Q: h.Q - other.Q,
		R: h.R - other.R,
	}
}

// Scale multiplies this coordinate by a scalar
func (h HexCoord) Scale(scalar int) HexCoord {
	return HexCoord{
		Q: h.Q * scalar,
		R: h.R * scalar,
	}
}

// String returns a string representation of this coordinate
func (h HexCoord) String() string {
	return "(" + strconv.Itoa(h.Q) + "," + strconv.Itoa(h.R) + ")"
}

// HexDirection represents the 6 directions in a hexagonal grid
type HexDirection int

const (
	HexDirectionEast HexDirection = iota
	HexDirectionNorthEast
	HexDirectionNorthWest
	HexDirectionWest
	HexDirectionSouthWest
	HexDirectionSouthEast
)

// HexDirections contains all 6 hexagonal directions
var HexDirections = []HexDirection{
	HexDirectionEast,
	HexDirectionNorthEast,
	HexDirectionNorthWest,
	HexDirectionWest,
	HexDirectionSouthWest,
	HexDirectionSouthEast,
}

// DQ returns the Q coordinate delta for this direction
func (d HexDirection) DQ() int {
	switch d {
	case HexDirectionEast:
		return 1
	case HexDirectionNorthEast:
		return 1
	case HexDirectionNorthWest:
		return 0
	case HexDirectionWest:
		return -1
	case HexDirectionSouthWest:
		return -1
	case HexDirectionSouthEast:
		return 0
	default:
		return 0
	}
}

// DR returns the R coordinate delta for this direction
func (d HexDirection) DR() int {
	switch d {
	case HexDirectionEast:
		return 0
	case HexDirectionNorthEast:
		return -1
	case HexDirectionNorthWest:
		return -1
	case HexDirectionWest:
		return 0
	case HexDirectionSouthWest:
		return 1
	case HexDirectionSouthEast:
		return 1
	default:
		return 0
	}
}

// Opposite returns the opposite direction
func (d HexDirection) Opposite() HexDirection {
	return HexDirection((int(d) + 3) % 6)
}

// RotateLeft returns the direction rotated 60 degrees counter-clockwise
func (d HexDirection) RotateLeft() HexDirection {
	return HexDirection((int(d) + 5) % 6)
}

// RotateRight returns the direction rotated 60 degrees clockwise
func (d HexDirection) RotateRight() HexDirection {
	return HexDirection((int(d) + 1) % 6)
}

// HexGrid represents a hexagonal grid
type HexGrid struct {
	HexSize matrix.Float
	Origin  HexCoord
}

// NewHexGrid creates a new hexagonal grid
func NewHexGrid(hexSize matrix.Float, origin HexCoord) *HexGrid {
	return &HexGrid{
		HexSize: hexSize,
		Origin:  origin,
	}
}

// ToWorld converts grid coordinates to world coordinates
func (g *HexGrid) ToWorld(coord HexCoord) matrix.Vec3 {
	return coord.ToWorld(g.HexSize)
}

// ToWorldWithHeight converts grid coordinates to world coordinates with height
func (g *HexGrid) ToWorldWithHeight(coord HexCoord, height matrix.Float) matrix.Vec3 {
	return coord.ToWorldWithHeight(g.HexSize, height)
}

// FromWorld converts world coordinates to the nearest hexagonal coordinate
func (g *HexGrid) FromWorld(worldPos matrix.Vec3) HexCoord {
	x := worldPos.X()
	z := worldPos.Z()

	q := matrix.Round(matrix.Float(2.0/3.0) * (x / g.HexSize))
	r := matrix.Round((matrix.Float(-1.0/3.0)*x + matrix.Float(math.Sqrt(3)/3.0)*z) / g.HexSize)

	return HexCoord{
		Q: int(q),
		R: int(r),
	}
}

// GetBounds returns the world bounds for a rectangular region of hexagonal coordinates
func (g *HexGrid) GetBounds(minCoord, maxCoord HexCoord) WorldBounds {
	minWorld := g.ToWorld(minCoord)
	maxWorld := g.ToWorld(maxCoord)

	// Add hex size padding
	padding := g.HexSize

	return WorldBounds{
		Min: matrix.NewVec3(
			minWorld.X()-padding,
			minWorld.Y()-padding,
			minWorld.Z()-padding,
		),
		Max: matrix.NewVec3(
			maxWorld.X()+padding,
			maxWorld.Y()+padding,
			maxWorld.Z()+padding,
		),
	}
}

// GetRegion returns all hexagonal coordinates in a rectangular region
func (g *HexGrid) GetRegion(minCoord, maxCoord HexCoord) []HexCoord {
	coords := make([]HexCoord, 0)

	for q := minCoord.Q; q <= maxCoord.Q; q++ {
		for r := minCoord.R; r <= maxCoord.R; r++ {
			coords = append(coords, HexCoord{Q: q, R: r})
		}
	}

	return coords
}

// WorldBounds represents axis-aligned world bounds
type WorldBounds struct {
	Min matrix.Vec3
	Max matrix.Vec3
}

// Contains checks if a world position is within these bounds
func (wb WorldBounds) Contains(pos matrix.Vec3) bool {
	return pos.X() >= wb.Min.X() && pos.X() <= wb.Max.X() &&
		pos.Y() >= wb.Min.Y() && pos.Y() <= wb.Max.Y() &&
		pos.Z() >= wb.Min.Z() && pos.Z() <= wb.Max.Z()
}

// GetCenter returns the center of these bounds
func (wb WorldBounds) GetCenter() matrix.Vec3 {
	return matrix.NewVec3(
		(wb.Min.X()+wb.Max.X())/2,
		(wb.Min.Y()+wb.Max.Y())/2,
		(wb.Min.Z()+wb.Max.Z())/2,
	)
}

// GetSize returns the size of these bounds
func (wb WorldBounds) GetSize() matrix.Vec3 {
	return matrix.NewVec3(
		wb.Max.X()-wb.Min.X(),
		wb.Max.Y()-wb.Min.Y(),
		wb.Max.Z()-wb.Min.Z(),
	)
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
