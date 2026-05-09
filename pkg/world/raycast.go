package world

import (
	"math"

	"github.com/tesselstudio/TesselBox/pkg/types"
)

// RaycastHit represents a raycast hit result
type RaycastHit struct {
	Hit        bool
	BlockPos   Vec3
	FaceNormal Vec3
	Distance   float32
	BlockID    BlockID
	ChunkCoord ChunkCoord
}

// Raycast casts a ray through the world and returns the hit block
func (w *World) Raycast(origin, direction types.Vec3, maxDistance float32) *RaycastHit {
	w.mu.RLock()
	defer w.mu.RUnlock()

	hit := &RaycastHit{Hit: false}

	// Step through the ray in small increments
	stepSize := float32(0.1) // 10cm steps
	steps := int(maxDistance / stepSize)

	currentPos := origin
	closestHit := float32(maxDistance + 1)

	for step := 0; step < steps; step++ {
		// Check current position
		blockX := int(math.Round(float64(currentPos.X)))
		blockY := int(math.Round(float64(currentPos.Y)))
		blockZ := int(math.Round(float64(currentPos.Z)))

		// Get block at this position
		block := w.GetBlock(blockX, blockY, blockZ)
		if block.ID != 0 { // Non-air block hit
			distance := currentPos.Distance(origin)
			if distance < closestHit {
				closestHit = distance
				hit.Hit = true
				hit.BlockPos = NewVec3(float32(blockX), float32(blockY), float32(blockZ))
				hit.BlockID = block.ID
				hit.Distance = distance

				// Determine face normal (which face was hit)
				dx := currentPos.X - float32(blockX)
				dy := currentPos.Y - float32(blockY)
				dz := currentPos.Z - float32(blockZ)

				absDX := math.Abs(float64(dx))
				absDY := math.Abs(float64(dy))
				absDZ := math.Abs(float64(dz))

				if absDX > absDY && absDX > absDZ {
					if dx > 0 {
						hit.FaceNormal = NewVec3(1, 0, 0) // Right face
					} else {
						hit.FaceNormal = NewVec3(-1, 0, 0) // Left face
					}
				} else if absDY > absDX && absDY > absDZ {
					if dy > 0 {
						hit.FaceNormal = NewVec3(0, 1, 0) // Top face
					} else {
						hit.FaceNormal = NewVec3(0, -1, 0) // Bottom face
					}
				} else {
					if dz > 0 {
						hit.FaceNormal = NewVec3(0, 0, 1) // Front face
					} else {
						hit.FaceNormal = NewVec3(0, 0, -1) // Back face
					}
				}
			}
		}

		// Move to next position
		currentPos = types.NewVec3(
			currentPos.X+direction.X*stepSize,
			currentPos.Y+direction.Y*stepSize,
			currentPos.Z+direction.Z*stepSize,
		)
	}

	return hit
}

// GetSafeSpawnHeight returns a safe Y position to spawn at given X, Z
func (w *World) GetSafeSpawnHeight(x, z int) int {
	// Ensure the chunk at spawn position is loaded and generated
	chunk := w.chunkManager.GetChunkByWorld(x, z)
	if chunk == nil {
		// Force load and generate the chunk
		coord, _, _, _ := WorldToLocal(x, 0, z)
		chunk = w.chunkManager.createChunk(coord)
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	// Start from top and work down to find ground
	groundY := -1
	for y := 255; y >= 0; y-- {
		block := w.GetBlock(x, y, z)
		if block.ID != 0 { // Found solid block
			groundY = y
			break
		}
	}

	if groundY == -1 {
		return 70 // Default height if nothing found
	}

	// Spawn player high in sky and let them fall naturally
	skyY := 200 // High spawn point in sky

	println("🔍 SPAWN DEBUG: Ground Y:", groundY, "Spawning at sky Y:", skyY)

	// No need to check for obstructions - player will fall through air
	println("🔍 SPAWN DEBUG: Final sky spawn Y:", skyY, "Position (", x, ",", skyY, ",", z, ")")
	return skyY
}
