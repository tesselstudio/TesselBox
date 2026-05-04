package audio

import (
	"math"
)

// CalculateAttenuation calculates volume attenuation based on distance
func CalculateAttenuation(listenerPos, soundPos Vec3) float32 {
	dx := soundPos.X - listenerPos.X
	dy := soundPos.Y - listenerPos.Y
	dz := soundPos.Z - listenerPos.Z

	distance := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))

	// Inverse square law with max distance
	maxDistance := float32(32.0) // Maximum audible distance in blocks
	minDistance := float32(1.0)  // Distance at which sound is at full volume

	if distance <= minDistance {
		return 1.0
	}

	if distance >= maxDistance {
		return 0.0
	}

	// Smooth attenuation curve
	ratio := (distance - minDistance) / (maxDistance - minDistance)
	attenuation := 1.0 - (ratio * ratio)

	return attenuation
}

// CalculatePan calculates stereo panning (-1.0 left to 1.0 right)
func CalculatePan(listenerPos, soundPos Vec3) float32 {
	dx := soundPos.X - listenerPos.X
	dz := soundPos.Z - listenerPos.Z

	// Calculate angle relative to listener's forward direction
	angle := float32(math.Atan2(float64(dx), float64(dz)))

	// Normalize angle to -1.0 to 1.0 range
	// atan2 returns -pi to pi, we want to map this to pan
	pan := float32(math.Sin(float64(angle)))

	return clamp(pan, -1.0, 1.0)
}

// Distance calculates the distance between two 3D points
func Distance(a, b Vec3) float32 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}
