package survival

import (
	"time"

	"github.com/tesselstudio/TesselBox/pkg/types"
	"github.com/tesselstudio/TesselBox/pkg/world"
)

// EnvironmentalHazards manages damage from environmental sources
type EnvironmentalHazards struct {
	// Damage counters
	lastLavaDamage time.Time
	lastFireDamage time.Time
	lastDrownTime  time.Time

	// Timers
	lavaDamageInterval time.Duration
	fireDamageInterval time.Duration
	drownInterval      time.Duration
}

// NewEnvironmentalHazards creates a new environmental hazards manager
func NewEnvironmentalHazards() *EnvironmentalHazards {
	return &EnvironmentalHazards{
		lastLavaDamage:     time.Now(),
		lastFireDamage:     time.Now(),
		lastDrownTime:      time.Now(),
		lavaDamageInterval: 500 * time.Millisecond,  // 2 damage per second in lava
		fireDamageInterval: 500 * time.Millisecond,  // 2 damage per second in fire
		drownInterval:      500 * time.Millisecond,  // Drown every 0.5 seconds
	}
}

// CheckEnvironmentalDamage checks for and applies environmental damage
func (eh *EnvironmentalHazards) CheckEnvironmentalDamage(playerPos types.Vec3, worldRef *world.World, stats *PlayerStats) {
	if worldRef == nil || stats == nil {
		return
	}

	blockX := int(playerPos.X)
	blockY := int(playerPos.Y)
	blockZ := int(playerPos.Z)

	// Get block at player position
	blockAtFeet := worldRef.GetBlock(blockX, blockY, blockZ)
	blockAtEyes := worldRef.GetBlock(blockX, blockY+1, blockZ)

	// Check for lava damage
	if blockAtFeet.ID == 11 || blockAtEyes.ID == 11 { // Lava block ID
		eh.applyLavaDamage(stats)
	}

	// Check for fire damage
	if blockAtFeet.ID == 12 || blockAtEyes.ID == 12 { // Fire block ID
		eh.applyFireDamage(stats)
	}

	// Check for drowning in water
	if blockAtEyes.ID == 10 { // Water block ID
		eh.applyDrowningDamage(stats)
	} else {
		// Reset drown timer when not underwater
		eh.lastDrownTime = time.Now()
	}

	// Check for suffocation (stuck in solid block)
	if blockAtEyes.ID > 0 && blockAtEyes.ID != 10 && blockAtEyes.ID != 11 && blockAtEyes.ID != 12 {
		eh.applySuffocationDamage(stats)
	}
}

// applyLavaDamage applies damage from lava
func (eh *EnvironmentalHazards) applyLavaDamage(stats *PlayerStats) {
	now := time.Now()
	if now.Sub(eh.lastLavaDamage) < eh.lavaDamageInterval {
		return
	}

	eh.lastLavaDamage = now

	// Apply fire effect and damage
	stats.Health.TakeDamage(2.0) // 2 HP per damage interval = 4 HP/sec
	stats.AddEffect("burning", 8.0)
}

// applyFireDamage applies damage from fire
func (eh *EnvironmentalHazards) applyFireDamage(stats *PlayerStats) {
	now := time.Now()
	if now.Sub(eh.lastFireDamage) < eh.fireDamageInterval {
		return
	}

	eh.lastFireDamage = now

	// Apply burning effect and damage
	stats.Health.TakeDamage(1.0) // 1 HP per damage interval = 2 HP/sec
	stats.AddEffect("burning", 8.0)
}

// applyDrowningDamage applies drowning damage
func (eh *EnvironmentalHazards) applyDrowningDamage(stats *PlayerStats) {
	now := time.Now()
	if now.Sub(eh.lastDrownTime) < eh.drownInterval {
		return
	}

	eh.lastDrownTime = now

	// Drown damage
	stats.Health.TakeDamage(2.0) // 2 HP per interval
	stats.AddEffect("drowning", 1.0)
}

// applySuffocationDamage applies damage from suffocation
func (eh *EnvironmentalHazards) applySuffocationDamage(stats *PlayerStats) {
	// Suffocation is instant death after extended time
	// We'll apply slow damage instead
	stats.Health.TakeDamage(0.5)
}

// GetHazardStatus returns a string describing current hazards
func (eh *EnvironmentalHazards) GetHazardStatus(playerPos types.Vec3, worldRef *world.World) string {
	if worldRef == nil {
		return ""
	}

	blockX := int(playerPos.X)
	blockY := int(playerPos.Y)
	blockZ := int(playerPos.Z)

	blockAtFeet := worldRef.GetBlock(blockX, blockY, blockZ)
	blockAtEyes := worldRef.GetBlock(blockX, blockY+1, blockZ)

	if blockAtFeet.ID == 11 || blockAtEyes.ID == 11 {
		return "IN LAVA"
	}
	if blockAtFeet.ID == 12 || blockAtEyes.ID == 12 {
		return "ON FIRE"
	}
	if blockAtEyes.ID == 10 {
		return "DROWNING"
	}
	if blockAtEyes.ID > 0 && blockAtEyes.ID != 10 && blockAtEyes.ID != 11 && blockAtEyes.ID != 12 {
		return "SUFFOCATING"
	}

	return ""
}

// ResetTimers resets all hazard timers
func (eh *EnvironmentalHazards) ResetTimers() {
	eh.lastLavaDamage = time.Now()
	eh.lastFireDamage = time.Now()
	eh.lastDrownTime = time.Now()
}

// VoidDamage checks if player is below world and applies void damage
func VoidDamage(playerPos types.Vec3, stats *PlayerStats) bool {
	if playerPos.Y < -64 {
		stats.Health.TakeDamage(20.0) // Instant death in void
		return true
	}
	return false
}

// FallDamage calculates and applies fall damage
func FallDamage(fallDistance float32, stats *PlayerStats) {
	if fallDistance < 3.0 {
		return // No damage from short falls (less than 3 blocks)
	}

	// Calculate damage based on fall distance
	// First 3 blocks are safe, then 0.5 damage per block
	damage := (fallDistance - 3.0) * 0.5

	if damage > 0 {
		stats.Health.TakeDamage(damage)
	}
}

// GetEnvironmentalTemperature estimates temperature at a location
// Useful for future biome/environment features
func GetEnvironmentalTemperature(worldRef *world.World, playerPos types.Vec3) float32 {
	if worldRef == nil {
		return 20.0 // Default temperature
	}

	// Base temperature
	temperature := float32(20.0)

	// Altitude affects temperature (higher = colder)
	if playerPos.Y > 100 {
		temperature -= (playerPos.Y - 100) * 0.01
	}

	// Check for fire/lava nearby (would increase temperature)
	blockAtFeet := worldRef.GetBlock(int(playerPos.X), int(playerPos.Y), int(playerPos.Z))
	if blockAtFeet.ID == 11 || blockAtFeet.ID == 12 {
		temperature += 20.0 // Very hot in lava/fire
	}

	return temperature
}
