package player

import (
	"sync"
	"time"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

// RemotePlayer represents another player in multiplayer
type RemotePlayer struct {
	mu sync.RWMutex

	// Player info
	ID   uint32
	Name string

	// Entity
	entity *engine.Entity

	// Visual components
	nametag *rendering.Drawing
	body    *rendering.Drawing

	// Position (interpolated)
	currentPos matrix.Vec3
	targetPos  matrix.Vec3
	rotation   matrix.Vec3

	// Interpolation
	lastUpdate time.Time
	updateRate float32

	// Host reference
	host *engine.Host
}

// RemotePlayerManager manages all remote player entities
type RemotePlayerManager struct {
	mu sync.RWMutex

	host          *engine.Host
	remotePlayers map[uint32]*RemotePlayer
}

// NewRemotePlayerManager creates a new manager for remote players
func NewRemotePlayerManager(host *engine.Host) *RemotePlayerManager {
	return &RemotePlayerManager{
		host:          host,
		remotePlayers: make(map[uint32]*RemotePlayer),
	}
}

// CreatePlayer creates a visual representation of a remote player
func (rpm *RemotePlayerManager) CreatePlayer(playerID uint32, name string, position matrix.Vec3) *RemotePlayer {
	rpm.mu.Lock()
	defer rpm.mu.Unlock()

	// Check if player already exists
	if existing, ok := rpm.remotePlayers[playerID]; ok {
		existing.SetPosition(position)
		return existing
	}

	player := &RemotePlayer{
		ID:         playerID,
		Name:       name,
		host:       rpm.host,
		currentPos: position,
		targetPos:  position,
		lastUpdate: time.Now(),
		updateRate: 20.0, // 20 updates per second
	}

	// Create entity
	player.entity = engine.NewEntity(rpm.host.WorkGroup())
	player.entity.Transform.SetPosition(position)

	// Create visual representation (simple colored cube for now)
	player.createVisuals()

	rpm.remotePlayers[playerID] = player
	return player
}

// RemovePlayer removes a remote player
func (rpm *RemotePlayerManager) RemovePlayer(playerID uint32) {
	rpm.mu.Lock()
	defer rpm.mu.Unlock()

	if player, ok := rpm.remotePlayers[playerID]; ok {
		player.Destroy()
		delete(rpm.remotePlayers, playerID)
	}
}

// GetPlayer returns a remote player by ID
func (rpm *RemotePlayerManager) GetPlayer(playerID uint32) *RemotePlayer {
	rpm.mu.RLock()
	defer rpm.mu.RUnlock()
	return rpm.remotePlayers[playerID]
}

// Update updates all remote players (call every frame)
func (rpm *RemotePlayerManager) Update(deltaTime float32) {
	rpm.mu.RLock()
	players := make([]*RemotePlayer, 0, len(rpm.remotePlayers))
	for _, player := range rpm.remotePlayers {
		players = append(players, player)
	}
	rpm.mu.RUnlock()

	for _, player := range players {
		player.Update(deltaTime)
	}
}

// GetAllPlayers returns all remote players
func (rpm *RemotePlayerManager) GetAllPlayers() map[uint32]*RemotePlayer {
	rpm.mu.RLock()
	defer rpm.mu.RUnlock()

	players := make(map[uint32]*RemotePlayer)
	for id, player := range rpm.remotePlayers {
		players[id] = player
	}
	return players
}

// createVisuals creates the visual representation of the player
func (rp *RemotePlayer) createVisuals() {
	// Create player body (simple cube - scaled to player size)
	bodyMesh := rendering.NewMeshCube(rp.host.MeshCache())

	// Create shader data
	sd := shader_data_registry.Create("basic")
	if basic, ok := sd.(*shader_data_registry.ShaderDataStandard); ok {
		// Different color for remote players
		basic.Color = matrix.Color{0.2, 0.6, 1.0, 1.0}
	}

	// Get material
	mat, err := rp.host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		return
	}

	// Get texture
	tex, _ := rp.host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)

	// Create body drawing
	rp.body = &rendering.Drawing{
		Material:   mat.CreateInstance([]*rendering.Texture{tex}),
		Mesh:       bodyMesh,
		ShaderData: sd,
		Transform:  &rp.entity.Transform,
		ViewCuller: &rp.host.Cameras.Primary,
	}

	rp.host.Drawings.AddDrawing(*rp.body)

	// Cleanup on destroy
	rp.entity.OnDestroy.Add(func() {
		if rp.body != nil {
			sd.Destroy()
		}
	})
}

// Update interpolates the player position
func (rp *RemotePlayer) Update(deltaTime float32) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	// Interpolate position towards target
	t := deltaTime * rp.updateRate
	t = min(t, 1.0) // Clamp to 1.0

	rp.currentPos = matrix.Vec3Lerp(rp.currentPos, rp.targetPos, t)

	// Update entity transform
	if rp.entity != nil {
		rp.entity.Transform.SetPosition(rp.currentPos)
		rp.entity.Transform.SetRotation(rp.rotation)
	}
}

// SetPosition updates the target position (called from network update)
func (rp *RemotePlayer) SetPosition(position matrix.Vec3) {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.targetPos = position
	rp.lastUpdate = time.Now()
}

// SetRotation updates the rotation
func (rp *RemotePlayer) SetRotation(rotation matrix.Vec3) {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.rotation = rotation
}

// GetPosition returns the current interpolated position
func (rp *RemotePlayer) GetPosition() matrix.Vec3 {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.currentPos
}

// GetRotation returns the rotation
func (rp *RemotePlayer) GetRotation() matrix.Vec3 {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.rotation
}

// Destroy cleans up the remote player
func (rp *RemotePlayer) Destroy() {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if rp.entity != nil {
		// Deactivate the entity (can't directly destroy from outside package)
		rp.entity.SetActive(false)
		rp.entity = nil
	}

	rp.body = nil
	rp.nametag = nil
}

// IsStale returns true if we haven't received updates recently
func (rp *RemotePlayer) IsStale() bool {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return time.Since(rp.lastUpdate) > 5*time.Second
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
