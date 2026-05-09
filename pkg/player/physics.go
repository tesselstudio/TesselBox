package player

import (
	"math"

	"github.com/tesselstudio/TesselBox/pkg/types"
	"github.com/tesselstudio/TesselBox/pkg/world"
)

// AABB represents an axis-aligned bounding box for collision detection
type AABB struct {
	Min types.Vec3
	Max types.Vec3
}

// NewAABB creates a new axis-aligned bounding box
func NewAABB(min, max types.Vec3) *AABB {
	return &AABB{Min: min, Max: max}
}

// Intersects checks if this AABB intersects with another AABB
func (a *AABB) Intersects(b *AABB) bool {
	return a.Min.X < b.Max.X && a.Max.X > b.Min.X &&
		a.Min.Y < b.Max.Y && a.Max.Y > b.Min.Y &&
		a.Min.Z < b.Max.Z && a.Max.Z > b.Min.Z
}

// Contains checks if a point is inside this AABB
func (a *AABB) Contains(point types.Vec3) bool {
	return point.X >= a.Min.X && point.X <= a.Max.X &&
		point.Y >= a.Min.Y && point.Y <= a.Max.Y &&
		point.Z >= a.Min.Z && point.Z <= a.Max.Z
}

// GetBounds returns the player's bounding box
func (p *Player) GetBounds() *AABB {
	width := float32(0.3)  // Player width
	height := float32(1.8) // Player height
	return NewAABB(
		types.NewVec3(p.position.X-width/2, p.position.Y-height/2, p.position.Z-width/2),
		types.NewVec3(p.position.X+width/2, p.position.Y+height/2, p.position.Z+width/2),
	)
}

// resolveCollisions resolves collisions with the world
func (p *Player) resolveCollisions(newPos types.Vec3) types.Vec3 {
	if p.worldRef == nil {
		return newPos
	}

	// Get nearby blocks (only check blocks within reach)
	checkDistance := float32(2.0)
	minX := int(math.Floor(float64(newPos.X - checkDistance)))
	maxX := int(math.Ceil(float64(newPos.X + checkDistance)))
	minY := int(math.Floor(float64(newPos.Y - checkDistance)))
	maxY := int(math.Ceil(float64(newPos.Y + checkDistance)))
	minZ := int(math.Floor(float64(newPos.Z - checkDistance)))
	maxZ := int(math.Ceil(float64(newPos.Z + checkDistance)))

	// Create player AABB at new position
	width := float32(0.3)
	height := float32(1.8)
	playerAABB := NewAABB(
		types.NewVec3(newPos.X-width/2, newPos.Y-height/2, newPos.Z-width/2),
		types.NewVec3(newPos.X+width/2, newPos.Y+height/2, newPos.Z+width/2),
	)

	// Check for collisions
	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			for z := minZ; z <= maxZ; z++ {
				block := p.worldRef.GetBlock(x, y, z)
				if block.ID == 0 { // Air block
					continue
				}

				// Create AABB for block
				blockAABB := NewAABB(
					types.NewVec3(float32(x)-0.5, float32(y)-0.5, float32(z)-0.5),
					types.NewVec3(float32(x)+0.5, float32(y)+0.5, float32(z)+0.5),
				)

				if playerAABB.Intersects(blockAABB) {

					// Resolve collision by pushing player out
					// Find the axis of least penetration
					dx := minOverlap(playerAABB.Min.X, playerAABB.Max.X, blockAABB.Min.X, blockAABB.Max.X)
					dy := minOverlap(playerAABB.Min.Y, playerAABB.Max.Y, blockAABB.Min.Y, blockAABB.Max.Y)
					dz := minOverlap(playerAABB.Min.Z, playerAABB.Max.Z, blockAABB.Min.Z, blockAABB.Max.Z)

					// Resolve on the axis with least penetration
					if math.Abs(float64(dx)) < math.Abs(float64(dy)) && math.Abs(float64(dx)) < math.Abs(float64(dz)) {
						// X axis
						if dx > 0 {
							newPos.X = blockAABB.Max.X + width/2 + 0.001
						} else {
							newPos.X = blockAABB.Min.X - width/2 - 0.001
						}
						p.velocity.X = 0 // Stop velocity in that direction
					} else if math.Abs(float64(dy)) < math.Abs(float64(dz)) {
						// Y axis
						if dy > 0 {
							newPos.Y = blockAABB.Max.Y + height/2 + 0.001
							p.velocity.Y = 0
							p.onGround = true
							p.jumping = false
						} else {
							newPos.Y = blockAABB.Min.Y - height/2 - 0.001
							p.velocity.Y = 0 // Hit head
						}
					} else {
						// Z axis
						if dz > 0 {
							newPos.Z = blockAABB.Max.Z + width/2 + 0.001
						} else {
							newPos.Z = blockAABB.Min.Z - width/2 - 0.001
						}
						p.velocity.Z = 0
					}
				}
			}
		}
	}

	return newPos
}

// minOverlap calculates the minimum overlap between two intervals
func minOverlap(min1, max1, min2, max2 float32) float32 {
	if max1 < min2 || max2 < min1 {
		return 0 // No overlap
	}

	// Calculate overlap amounts
	overlapLeft := max2 - min1
	overlapRight := max1 - min2

	// Return minimum overlap with sign indicating direction
	if overlapLeft < overlapRight {
		return overlapLeft
	}
	return -overlapRight
}

// checkOnGround checks if the player is standing on solid ground
func (p *Player) checkOnGround() bool {
	if p.worldRef == nil {
		return false
	}

	// Check block below player feet
	checkDistance := float32(0.1)
	blockY := int(math.Floor(float64(p.position.Y - checkDistance)))

	// Get nearby X and Z positions
	checkRadius := 0.2
	for dx := -checkRadius; dx <= checkRadius; dx += 0.2 {
		for dz := -checkRadius; dz <= checkRadius; dz += 0.2 {
			blockX := int(math.Round(float64(p.position.X + float32(dx))))
			blockZ := int(math.Round(float64(p.position.Z + float32(dz))))

			block := p.worldRef.GetBlock(blockX, blockY, blockZ)
			if block.ID != 0 { // Found solid ground
				return true
			}
		}
	}

	return false
}

// ApplyFallDamage applies damage based on fall distance
func (p *Player) ApplyFallDamage(fallDistance float32) {
	if fallDistance < 3.0 {
		return // No damage from short falls
	}

	// Damage = distance - safe fall distance
	damageAmount := int(fallDistance - 3.0)
	if damageAmount > 0 {
		p.stats.Health.TakeDamage(float32(damageAmount))
	}
}

// CheckBlockCollision checks if a specific block type exists at player position
func (p *Player) CheckBlockCollision(blockID world.BlockID) bool {
	if p.worldRef == nil {
		return false
	}

	// Get nearby blocks around player
	checkRadius := float32(0.5)
	minX := int(math.Floor(float64(p.position.X - checkRadius)))
	maxX := int(math.Ceil(float64(p.position.X + checkRadius)))
	minY := int(math.Floor(float64(p.position.Y - checkRadius)))
	maxY := int(math.Ceil(float64(p.position.Y + checkRadius)))
	minZ := int(math.Floor(float64(p.position.Z - checkRadius)))
	maxZ := int(math.Ceil(float64(p.position.Z + checkRadius)))

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			for z := minZ; z <= maxZ; z++ {
				block := p.worldRef.GetBlock(x, y, z)
				if block.ID == blockID {
					return true
				}
			}
		}
	}

	return false
}

// IsInWater checks if player is in water
func (p *Player) IsInWater() bool {
	// Check if player is submerged in water (BlockID 10)
	return p.CheckBlockCollision(world.BlockID(10))
}

// IsInLava checks if player is in lava
func (p *Player) IsInLava() bool {
	// Check if player is in lava (BlockID 11)
	return p.CheckBlockCollision(world.BlockID(11))
}

// IsInFire checks if player is in fire
func (p *Player) IsInFire() bool {
	// Check if player is in fire (BlockID 12)
	return p.CheckBlockCollision(world.BlockID(12))
}

// ApplyWaterPhysics applies water physics (friction, buoyancy)
func (p *Player) ApplyWaterPhysics(dt float32) {
	if !p.IsInWater() {
		return
	}

	// Apply water friction (reduce velocity)
	frictionMultiplier := float32(0.8)
	p.velocity.X *= frictionMultiplier
	p.velocity.Y *= frictionMultiplier
	p.velocity.Z *= frictionMultiplier

	// Apply buoyancy (counteract gravity)
	p.velocity.Y += p.gravity * 0.5 * dt
}

// ApplyLavaPhysics applies lava physics (high friction, slow movement)
func (p *Player) ApplyLavaPhysics(dt float32) {
	if !p.IsInLava() {
		return
	}

	// Lava has very high friction
	frictionMultiplier := float32(0.4)
	p.velocity.X *= frictionMultiplier
	p.velocity.Y *= frictionMultiplier
	p.velocity.Z *= frictionMultiplier

	// Apply some buoyancy but player still sinks
	p.velocity.Y += p.gravity * 0.1 * dt

	// Apply damage over time
	p.stats.Health.TakeDamage(4.0 * dt) // 4 hearts per second in lava
}
