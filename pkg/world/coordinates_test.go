package world

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestHexCoordToWorld(t *testing.T) {
	tests := []struct {
		name     string
		coord    HexCoord
		scale    float32
		expected matrix.Vec3
	}{
		{
			name:     "Origin",
			coord:    HexCoord{Q: 0, R: 0},
			scale:    1.0,
			expected: matrix.NewVec3(0, 0, 0),
		},
		{
			name:     "Positive Q",
			coord:    HexCoord{Q: 1, R: 0},
			scale:    1.0,
			expected: matrix.NewVec3(1.732, 0, 0),
		},
		{
			name:     "Positive R",
			coord:    HexCoord{Q: 0, R: 1},
			scale:    1.0,
			expected: matrix.NewVec3(0.866, 0, 1.5),
		},
		{
			name:     "Scale 2",
			coord:    HexCoord{Q: 1, R: 1},
			scale:    2.0,
			expected: matrix.NewVec3(5.196, 0, 3.0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.coord.ToWorld(tt.scale)

			// Allow small floating point errors
			tolerance := float32(0.01)
			if abs32(result.X()-tt.expected.X()) > tolerance {
				t.Errorf("X: expected %.3f, got %.3f", tt.expected.X(), result.X())
			}
			if abs32(result.Z()-tt.expected.Z()) > tolerance {
				t.Errorf("Z: expected %.3f, got %.3f", tt.expected.Z(), result.Z())
			}
		})
	}
}

func TestHexCoordDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        HexCoord
		b        HexCoord
		expected int
	}{
		{
			name:     "Same coordinate",
			a:        HexCoord{Q: 0, R: 0},
			b:        HexCoord{Q: 0, R: 0},
			expected: 0,
		},
		{
			name:     "Adjacent",
			a:        HexCoord{Q: 0, R: 0},
			b:        HexCoord{Q: 1, R: 0},
			expected: 1,
		},
		{
			name:     "Distance 2",
			a:        HexCoord{Q: 0, R: 0},
			b:        HexCoord{Q: 2, R: 0},
			expected: 2,
		},
		{
			name:     "Diagonal",
			a:        HexCoord{Q: 0, R: 0},
			b:        HexCoord{Q: 1, R: 1},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.Distance(tt.b)
			if result != tt.expected {
				t.Errorf("Expected distance %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestHexCoordNeighbor(t *testing.T) {
	center := HexCoord{Q: 0, R: 0}

	// Test all 6 directions
	directions := []HexDirection{
		HexDirectionEast,
		HexDirectionNorthEast,
		HexDirectionNorthWest,
		HexDirectionWest,
		HexDirectionSouthWest,
		HexDirectionSouthEast,
	}

	for _, dir := range directions {
		neighbor := center.Neighbor(dir)
		distance := center.Distance(neighbor)
		if distance != 1 {
			t.Errorf("Neighbor distance should be 1, got %d for direction %d", distance, dir)
		}
	}
}

func TestHexCoordString(t *testing.T) {
	coord := HexCoord{Q: 5, R: 3}
	expected := "(5,3)"
	result := coord.String()

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestHexCoordScale(t *testing.T) {
	coord := HexCoord{Q: 2, R: 3}
	scale := 3
	expected := HexCoord{Q: 6, R: 9}
	result := coord.Scale(scale)

	if result != expected {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
