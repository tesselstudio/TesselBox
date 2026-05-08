package types

import "math"

// Vec3 represents a 3D vector
type Vec3 struct {
	X, Y, Z float32
}

// NewVec3 creates a new Vec3
func NewVec3(x, y, z float32) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

// GetX returns the X component (for compatibility with matrix.Vec3 interface)
func (v Vec3) GetX() Float {
	return Float(v.X)
}

// GetY returns the Y component (for compatibility with matrix.Vec3 interface)
func (v Vec3) GetY() Float {
	return Float(v.Y)
}

// GetZ returns the Z component (for compatibility with matrix.Vec3 interface)
func (v Vec3) GetZ() Float {
	return Float(v.Z)
}

// Distance returns the distance between this vector and another vector
func (v Vec3) Distance(other Vec3) float32 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	dz := v.Z - other.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

// Vec2 represents a 2D vector
type Vec2 struct {
	X, Y float32
}

// NewVec2 creates a new Vec2
func NewVec2(x, y float32) Vec2 {
	return Vec2{X: x, Y: y}
}

// Color represents a simple RGBA color
type Color struct {
	R, G, B, A uint8
}

// NewColor creates a new color from RGBA values
func NewColor(r, g, b, a uint8) Color {
	return Color{R: r, G: g, B: b, A: a}
}

// ColorGray returns a gray color
func ColorGray() Color {
	return Color{128, 128, 128, 255}
}

// ColorWhite returns a white color
func ColorWhite() Color {
	return Color{255, 255, 255, 255}
}

// Float represents a simple float type
type Float float32

// Round rounds a Float to the nearest integer
func Round(f Float) int {
	if f < 0 {
		return int(f - 0.5)
	}
	return int(f + 0.5)
}

// SetY sets the Y component (for compatibility)
func (v *Vec3) SetY(y Float) {
	v.Y = float32(y)
}
