package player

import (
	"sync"
	"time"

	"github.com/tesselstudio/TesselBox/pkg/types"
)

// RemotePlayer represents another player in multiplayer (placeholder for pure OpenGL implementation)
type RemotePlayer struct {
	mu sync.RWMutex

	// Player info
	ID   uint32
	Name string

	// Transform
	position types.Vec3
	rotation types.Vec3

	// Physics
	velocity types.Vec3
	onGround bool

	// Animation
	lastUpdateTime time.Time
	currentAnim    string
	animTime       float32
}

// RemotePlayerManager manages all remote players (placeholder for pure OpenGL implementation)
type RemotePlayerManager struct {
	mu            sync.RWMutex
	remotePlayers map[uint32]*RemotePlayer
}

// NewRemotePlayerManager creates a new manager for remote players
func NewRemotePlayerManager() *RemotePlayerManager {
	return &RemotePlayerManager{
		remotePlayers: make(map[uint32]*RemotePlayer),
	}
}

// CreatePlayer creates a visual representation of a remote player
func (rpm *RemotePlayerManager) CreatePlayer(playerID uint32, name string, position types.Vec3) *RemotePlayer {
	rpm.mu.Lock()
	defer rpm.mu.Unlock()

	// Check if player already exists
	if _, exists := rpm.remotePlayers[playerID]; exists {
		return nil
	}

	// Create player (placeholder for pure OpenGL implementation)
	player := &RemotePlayer{
		ID:             playerID,
		Name:           name,
		position:       position,
		lastUpdateTime: time.Now(),
	}

	// Store player
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

// createPlayerVisual creates the visual representation of a remote player (placeholder for pure OpenGL implementation)
func (rpm *RemotePlayerManager) createPlayerVisual(player *RemotePlayer) {
	// Player visual creation will be implemented with pure OpenGL
	// For now, this is a placeholder
}

// Update interpolates the player position (placeholder for pure OpenGL implementation)
func (rp *RemotePlayer) Update(deltaTime float32) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	// Player position interpolation will be implemented with pure OpenGL
	// For now, this is a placeholder
	rp.lastUpdateTime = time.Now()
}

// SetPosition updates the target position (called from network update)
func (rp *RemotePlayer) SetPosition(position types.Vec3) {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.position = position
}

// SetRotation updates the rotation
func (rp *RemotePlayer) SetRotation(rotation types.Vec3) {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.rotation = rotation
}

// GetPosition returns the current interpolated position
func (rp *RemotePlayer) GetPosition() types.Vec3 {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.position
}

// GetRotation returns the rotation
func (rp *RemotePlayer) GetRotation() types.Vec3 {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.rotation
}

// Destroy cleans up the remote player (placeholder for pure OpenGL implementation)
func (rp *RemotePlayer) Destroy() {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	// Player cleanup will be implemented with pure OpenGL
	// For now, this is a placeholder
}

// IsStale returns true if we haven't received updates recently
func (rp *RemotePlayer) IsStale() bool {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return time.Since(rp.lastUpdateTime) > 5*time.Second
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
