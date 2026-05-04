package player

import (
	"math"
	"sync"
	"time"

	"github.com/tesselstudio/TesselBox/pkg/audio"
	"github.com/tesselstudio/TesselBox/pkg/crafting"
	"github.com/tesselstudio/TesselBox/pkg/effects"
	"github.com/tesselstudio/TesselBox/pkg/survival"
	"github.com/tesselstudio/TesselBox/pkg/world"
	"kaijuengine.com/engine"
	"kaijuengine.com/matrix"
)

// Player represents the local player in the game
type Player struct {
	mu sync.RWMutex

	// Entity reference
	entity *engine.Entity
	host   *engine.Host

	// Transform
	position matrix.Vec3
	rotation matrix.Vec3 // Pitch (X), Yaw (Y), Roll (Z)

	// Physics
	velocity matrix.Vec3
	onGround bool
	jumping  bool
	sneaking bool

	// Movement speeds
	walkSpeed   float32
	sprintSpeed float32
	sneakSpeed  float32
	jumpForce   float32
	gravity     float32

	// Survival stats
	stats *survival.PlayerStats

	// Inventory
	inventory  *crafting.Inventory
	hotbarSlot int

	// Interaction
	reachDistance float32
	selectedBlock world.BlockData

	// World reference
	worldRef *world.World

	// Effects manager
	effectManager *effects.EffectManager

	// Last update time
	lastUpdate time.Time

	// Footstep tracking
	lastFootstepTime time.Time
	lastPosition     matrix.Vec3
}

// NewPlayer creates a new player
func NewPlayer(host *engine.Host, worldRef *world.World) *Player {
	player := &Player{
		host:          host,
		worldRef:      worldRef,
		position:      matrix.NewVec3(0, 70, 0),
		rotation:      matrix.NewVec3(0, 0, 0),
		velocity:      matrix.NewVec3(0, 0, 0),
		walkSpeed:     4.3, // blocks per second
		sprintSpeed:   5.6, // blocks per second
		sneakSpeed:    1.3, // blocks per second
		jumpForce:     8.5,
		gravity:       28.0,
		reachDistance: 5.0,
		stats:         survival.NewPlayerStats(20, 20, 20),
		inventory:     crafting.NewInventory(36, 9), // 36 slots, 9 hotbar
		hotbarSlot:    0,
		selectedBlock: world.BlockData{ID: world.BlockID(1)}, // Stone
		lastUpdate:    time.Now(),
	}

	// Give starter items
	player.giveStarterItems()

	return player
}

// giveStarterItems gives the player starting items
func (p *Player) giveStarterItems() {
	itemRegistry := crafting.GetGlobalItemRegistry()

	starterItems := []struct {
		itemID   string
		quantity int
	}{
		{"wood", 16},
		{"stone", 32},
		{"dirt", 16},
		{"wooden_pickaxe", 1},
		{"wooden_sword", 1},
		{"apple", 8},
	}

	for i, starter := range starterItems {
		if i >= 36 {
			break
		}
		if item, exists := itemRegistry.GetItem(starter.itemID); exists {
			stack := crafting.NewItemStack(item, starter.quantity)
			p.inventory.SetSlot(i, stack)
		}
	}
}

// SetEntity sets the player's entity
func (p *Player) SetEntity(entity *engine.Entity) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.entity = entity
}

// GetEntity returns the player's entity
func (p *Player) GetEntity() *engine.Entity {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.entity
}

// Update updates the player (called every frame)
func (p *Player) Update(deltaTime float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	dt := float32(now.Sub(p.lastUpdate).Seconds())
	p.lastUpdate = now

	// Apply gravity
	if !p.onGround {
		p.velocity.SetY(p.velocity.Y() - p.gravity*dt)
	}

	// Apply velocity to position
	newPos := matrix.NewVec3(
		p.position.X()+p.velocity.X()*dt,
		p.position.Y()+p.velocity.Y()*dt,
		p.position.Z()+p.velocity.Z()*dt,
	)

	// Collision detection with world
	newPos = p.resolveCollisions(newPos)

	p.position = newPos

	// Check if on ground
	p.onGround = p.checkOnGround()

	// Apply friction
	p.velocity.SetX(p.velocity.X() * 0.9)
	p.velocity.SetZ(p.velocity.Z() * 0.9)

	if p.onGround {
		p.velocity.SetY(0)
	}

	// Update entity transform
	if p.entity != nil {
		p.entity.Transform.SetPosition(p.position)
	}

	// Update survival stats
	p.stats.Update(dt)
}

// GetPosition returns the player's position
func (p *Player) GetPosition() matrix.Vec3 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.position
}

// SetPosition sets the player's position
func (p *Player) SetPosition(pos matrix.Vec3) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.position = pos
	if p.entity != nil {
		p.entity.Transform.SetPosition(pos)
	}
}

// GetRotation returns the player's rotation (pitch, yaw, roll)
func (p *Player) GetRotation() matrix.Vec3 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.rotation
}

// SetRotation sets the player's rotation
func (p *Player) SetRotation(rot matrix.Vec3) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Clamp pitch to prevent flipping
	pitch := rot.X()
	if pitch > 89.0 {
		pitch = 89.0
	} else if pitch < -89.0 {
		pitch = -89.0
	}

	// Normalize yaw
	yaw := math.Mod(float64(rot.Y()), 360.0)

	p.rotation = matrix.NewVec3(pitch, float32(yaw), rot.Z())
}

// Move moves the player relative to their facing direction
func (p *Player) Move(forward, right float32, sprinting bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	speed := p.walkSpeed
	if sprinting && !p.sneaking {
		speed = p.sprintSpeed
	}
	if p.sneaking {
		speed = p.sneakSpeed
	}

	// Get yaw in radians
	yawRad := float64(p.rotation.Y()) * math.Pi / 180.0

	// Calculate movement vector
	moveX := (forward*float32(math.Sin(yawRad)) + right*float32(math.Cos(yawRad))) * speed
	moveZ := (forward*float32(math.Cos(yawRad)) - right*float32(math.Sin(yawRad))) * speed

	p.velocity.SetX(moveX)
	p.velocity.SetZ(moveZ)

	// Exhaust hunger when sprinting
	if sprinting {
		p.stats.Hunger.Exhaust(0.1)
	}
}

// Jump makes the player jump
func (p *Player) Jump() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.onGround && !p.jumping {
		p.velocity.SetY(p.jumpForce)
		p.jumping = true
		p.onGround = false

		// Exhaust hunger
		p.stats.Hunger.Exhaust(0.05)
	}
}

// StopJump resets jump state
func (p *Player) StopJump() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.jumping = false
}

// SetSneaking sets the sneaking state
func (p *Player) SetSneaking(sneaking bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sneaking = sneaking
}

// IsSneaking returns true if the player is sneaking
func (p *Player) IsSneaking() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.sneaking
}

// IsOnGround returns true if the player is on ground
func (p *Player) IsOnGround() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.onGround
}

// GetStats returns the player's survival stats
func (p *Player) GetStats() *survival.PlayerStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stats
}

// GetInventory returns the player's inventory
func (p *Player) GetInventory() *crafting.Inventory {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.inventory
}

// GetHotbarSlot returns the selected hotbar slot
func (p *Player) GetHotbarSlot() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.hotbarSlot
}

// SetHotbarSlot sets the selected hotbar slot
func (p *Player) SetHotbarSlot(slot int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if slot >= 0 && slot < 9 {
		p.hotbarSlot = slot
	}
}

// GetSelectedBlock returns the currently selected block for placement
func (p *Player) GetSelectedBlock() world.BlockData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.selectedBlock
}

// SetSelectedBlock sets the block to place
func (p *Player) SetSelectedBlock(block world.BlockData) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.selectedBlock = block
}

// GetReachDistance returns the block interaction reach distance
func (p *Player) GetReachDistance() float32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.reachDistance
}

// Raycast performs a raycast from the player's eyes
func (p *Player) Raycast(maxDistance float32) (hit bool, hitPos, hitNormal matrix.Vec3, hitBlock world.BlockData) {
	p.mu.RLock()
	pos := p.position
	rot := p.rotation
	p.mu.RUnlock()

	// Calculate ray direction from rotation
	pitchRad := float64(rot.X()) * math.Pi / 180.0
	yawRad := float64(rot.Y()) * math.Pi / 180.0

	dir := matrix.NewVec3(
		float32(-math.Sin(yawRad)*math.Cos(pitchRad)),
		float32(math.Sin(pitchRad)),
		float32(-math.Cos(yawRad)*math.Cos(pitchRad)),
	)

	// Raycast from eye position (slightly above player feet)
	eyePos := matrix.NewVec3(pos.X(), pos.Y()+1.6, pos.Z())

	// Simple voxel raycasting
	step := float32(0.1)
	for dist := float32(0); dist < maxDistance; dist += step {
		checkPos := matrix.NewVec3(
			eyePos.X()+dir.X()*dist,
			eyePos.Y()+dir.Y()*dist,
			eyePos.Z()+dir.Z()*dist,
		)

		blockX := int(math.Floor(float64(checkPos.X())))
		blockY := int(math.Floor(float64(checkPos.Y())))
		blockZ := int(math.Floor(float64(checkPos.Z())))

		block := p.worldRef.GetBlock(blockX, blockY, blockZ)
		if !block.IsAir() {
			// Calculate hit normal
			hitPos = checkPos
			return true, hitPos, dir.Normal(), block
		}
	}

	return false, matrix.NewVec3(0, 0, 0), matrix.NewVec3(0, 0, 0), world.BlockData{}
}

// PlaceBlock places a block from the hotbar at the targeted position
func (p *Player) PlaceBlock() bool {
	hit, hitPos, hitNormal, _ := p.Raycast(p.reachDistance)
	if !hit {
		return false
	}

	// Calculate placement position (adjacent to hit face)
	placeX := int(math.Floor(float64(hitPos.X() + hitNormal.X()*0.1)))
	placeY := int(math.Floor(float64(hitPos.Y() + hitNormal.Y()*0.1)))
	placeZ := int(math.Floor(float64(hitPos.Z() + hitNormal.Z()*0.1)))

	// Get block from hotbar slot
	p.mu.RLock()
	hotbarSlot := p.hotbarSlot
	p.mu.RUnlock()

	// Get item from hotbar
	hotbarItem := p.inventory.GetSlot(hotbarSlot)
	if hotbarItem == nil || hotbarItem.IsEmpty() {
		return false
	}

	// Convert item to block ID
	blockID := crafting.ItemIDToBlockID(hotbarItem.Item.ID)
	if blockID == 0 {
		return false // Not a placeable block
	}

	blockToPlace := world.BlockData{ID: world.BlockID(blockID)}

	// Check if position is occupied by player
	playerPos := p.position
	playerBlockX := int(math.Floor(float64(playerPos.X())))
	playerBlockY := int(math.Floor(float64(playerPos.Y())))
	playerBlockZ := int(math.Floor(float64(playerPos.Z())))

	// Don't place block inside player
	if placeX == playerBlockX && placeZ == playerBlockZ {
		if placeY == playerBlockY || placeY == playerBlockY+1 {
			return false
		}
	}

	// Place the block
	p.worldRef.SetBlock(placeX, placeY, placeZ, blockToPlace)

	// Play block placement sound
	audio.GetManager().TriggerGameEvent(audio.EventBlockPlace)

	// Consume one item from inventory
	hotbarItem.Quantity--
	if hotbarItem.Quantity <= 0 {
		p.inventory.SetSlot(hotbarSlot, nil)
	}

	return true
}

// BreakBlock breaks the targeted block and drops items
func (p *Player) BreakBlock() bool {
	hit, hitPos, _, hitBlock := p.Raycast(p.reachDistance)
	if !hit {
		return false
	}

	blockX := int(math.Floor(float64(hitPos.X())))
	blockY := int(math.Floor(float64(hitPos.Y())))
	blockZ := int(math.Floor(float64(hitPos.Z())))

	// Get the item that drops from this block
	itemID := crafting.BlockIDToItemID(uint16(hitBlock.ID))

	// Get block color for particles
	blockColor := p.getBlockColor(hitBlock.ID)

	// Remove the block (set to air)
	p.worldRef.SetBlock(blockX, blockY, blockZ, world.BlockData{ID: world.BlockID(0)}) // Air

	// Play block break sound
	audio.GetManager().TriggerGameEvent(audio.EventBlockBreak)

	// Spawn block break particles
	if p.effectManager != nil {
		p.effectManager.SpawnBlockBreak(hitPos, blockColor)
	}

	// Add to inventory if it's a valid item
	if itemID != "" {
		itemRegistry := crafting.GetGlobalItemRegistry()
		if item, exists := itemRegistry.GetItem(itemID); exists {
			stack := crafting.NewItemStack(item, 1)
			if err := p.inventory.AddItem(stack); err != nil {
				// Inventory full - could drop item on ground instead
			}
		}
	}

	return true
}

// getBlockColor returns the color for particle effects
func (p *Player) getBlockColor(id world.BlockID) matrix.Color {
	switch id {
	case world.BlockIDStone:
		return matrix.ColorGray()
	case world.BlockIDDirt:
		return matrix.NewColor(139.0/255.0, 90.0/255.0, 43.0/255.0, 1.0)
	case world.BlockIDGrass:
		return matrix.NewColor(124.0/255.0, 200.0/255.0, 50.0/255.0, 1.0)
	case world.BlockIDWood:
		return matrix.NewColor(139.0/255.0, 90.0/255.0, 43.0/255.0, 1.0)
	case world.BlockIDGlass:
		return matrix.NewColor(200.0/255.0, 200.0/255.0, 255.0/255.0, 0.6)
	default:
		return matrix.ColorWhite()
	}
}

// SetEffectManager sets the effect manager for particles
func (p *Player) SetEffectManager(em *effects.EffectManager) {
	p.effectManager = em
}

// UpdateFootsteps spawns footstep particles when walking
func (p *Player) UpdateFootsteps() {
	// Check if player is moving on ground
	if !p.onGround {
		return
	}

	// Check if moved enough distance
	dist := p.position.Distance(p.lastPosition)
	if dist < 0.5 {
		return
	}

	// Check time since last footstep
	if time.Since(p.lastFootstepTime) < 400*time.Millisecond {
		return
	}

	// Spawn footstep particles
	if p.effectManager != nil {
		footPos := matrix.NewVec3(p.position.X(), p.position.Y(), p.position.Z())
		p.effectManager.SpawnBlockBreak(footPos, matrix.NewColor(0.6, 0.5, 0.4, 0.5))
	}

	p.lastFootstepTime = time.Now()
	p.lastPosition = p.position
}

// resolveCollisions resolves collisions between player and world
func (p *Player) resolveCollisions(newPos matrix.Vec3) matrix.Vec3 {
	// Simple AABB collision resolution
	playerWidth := float32(0.6)
	playerHeight := float32(1.8)

	// Check horizontal collisions
	x := newPos.X()
	y := newPos.Y()
	z := newPos.Z()

	// Check X collision
	if p.checkCollision(x-playerWidth/2, p.position.Y(), p.position.Z(), playerWidth, playerHeight) ||
		p.checkCollision(x+playerWidth/2, p.position.Y(), p.position.Z(), playerWidth, playerHeight) {
		x = p.position.X()
		p.velocity.SetX(0)
	}

	// Check Z collision
	if p.checkCollision(x, p.position.Y(), z-playerWidth/2, playerWidth, playerHeight) ||
		p.checkCollision(x, p.position.Y(), z+playerWidth/2, playerWidth, playerHeight) {
		z = p.position.Z()
		p.velocity.SetZ(0)
	}

	// Check Y collision
	if p.checkCollision(x, y, z, playerWidth, playerHeight) {
		// Check if falling or jumping
		if p.velocity.Y() < 0 {
			// Hit ground
			for p.checkCollision(x, y, z, playerWidth, playerHeight) && y < 300 {
				y += 0.1
			}
		} else {
			// Hit ceiling
			for p.checkCollision(x, y, z, playerWidth, playerHeight) && y > -64 {
				y -= 0.1
			}
		}
		p.velocity.SetY(0)
	}

	return matrix.NewVec3(x, y, z)
}

// checkCollision checks if the player AABB intersects with any solid blocks
func (p *Player) checkCollision(x, y, z float32, width, height float32) bool {
	minX := int(math.Floor(float64(x - width/2)))
	maxX := int(math.Floor(float64(x + width/2)))
	minY := int(math.Floor(float64(y)))
	maxY := int(math.Floor(float64(y + height)))
	minZ := int(math.Floor(float64(z - width/2)))
	maxZ := int(math.Floor(float64(z + width/2)))

	for bx := minX; bx <= maxX; bx++ {
		for by := minY; by <= maxY; by++ {
			for bz := minZ; bz <= maxZ; bz++ {
				block := p.worldRef.GetBlock(bx, by, bz)
				if block.IsSolid() {
					return true
				}
			}
		}
	}

	return false
}

// checkOnGround checks if the player is standing on solid ground
func (p *Player) checkOnGround() bool {
	x := p.position.X()
	y := p.position.Y()
	z := p.position.Z()
	width := float32(0.6)

	// Check if there's solid ground 0.1 blocks below
	return p.checkCollision(x, y-0.1, z, width, 0.1)
}

// DropItem drops one item from the selected hotbar slot
func (p *Player) DropItem() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Get item from hotbar slot
	hotbarItem := p.inventory.GetSlot(p.hotbarSlot)
	if hotbarItem == nil || hotbarItem.IsEmpty() {
		return false
	}

	// Remove one item
	hotbarItem.Quantity--
	if hotbarItem.Quantity <= 0 {
		p.inventory.SetSlot(p.hotbarSlot, nil)
	}

	// TODO: Spawn dropped item entity in the world at player position
	// For now, the item is just removed from inventory

	return true
}

// Respawn respawns the player at the spawn point
func (p *Player) Respawn() {
	p.mu.Lock()
	defer p.mu.Unlock()

	spawn := p.worldRef.GetSpawnPoint()
	safeY := p.worldRef.GetSafeSpawnHeight(int(spawn.X()), int(spawn.Z()))

	p.position = matrix.NewVec3(spawn.X(), float32(safeY), spawn.Z())
	p.velocity = matrix.NewVec3(0, 0, 0)
	p.rotation = matrix.NewVec3(0, 0, 0)
	p.onGround = false

	// Reset stats
	p.stats.Reset()
}
