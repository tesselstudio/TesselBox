package effects

import (
	"math"
	"math/rand"
	"time"

	"github.com/tesselstudio/TesselBox/pkg/types"
)

// Particle represents a single particle
type Particle struct {
	Position types.Vec3
	Velocity types.Vec3
	Color    types.Color
	Size     float32
	Life     float32
	MaxLife  float32
	Active   bool
}

// ParticleSystem manages a collection of particles
type ParticleSystem struct {
	particles []Particle
}

// NewParticleSystem creates a new particle system
func NewParticleSystem(maxParticles int) *ParticleSystem {
	ps := &ParticleSystem{
		particles: make([]Particle, maxParticles),
	}

	return ps
}

// SpawnBlockBreakParticles creates particles when a block is broken
func (ps *ParticleSystem) SpawnBlockBreakParticles(pos types.Vec3, blockColor types.Color, count int) {
	for i := 0; i < count; i++ {
		// Find inactive particle
		for j := range ps.particles {
			if !ps.particles[j].Active {
				// Random velocity upward and outward
				angle := rand.Float64() * 2 * math.Pi
				speed := float32(rand.Float32()*2 + 1)
				vx := float32(math.Cos(angle)) * speed
				vz := float32(math.Sin(angle)) * speed
				vy := float32(rand.Float32()*3 + 2)

				ps.particles[j] = Particle{
					Position: pos,
					Velocity: types.NewVec3(vx, vy, vz),
					Color:    blockColor,
					Size:     float32(rand.Float32()*0.1 + 0.05),
					Life:     1.0,
					MaxLife:  1.0,
					Active:   true,
				}
				break
			}
		}
	}
}

// SpawnFootstepParticles creates dust particles when player walks
func (ps *ParticleSystem) SpawnFootstepParticles(pos types.Vec3, count int) {
	for i := 0; i < count; i++ {
		for j := range ps.particles {
			if !ps.particles[j].Active {
				// Random velocity upward
				vx := float32(rand.Float32()*0.5 - 0.25)
				vz := float32(rand.Float32()*0.5 - 0.25)
				vy := float32(rand.Float32() * 0.5)

				ps.particles[j] = Particle{
					Position: pos,
					Velocity: types.NewVec3(vx, vy, vz),
					Color:    types.NewColor(178, 178, 153, 204),
					Size:     float32(rand.Float32()*0.05 + 0.02),
					Life:     0.5,
					MaxLife:  0.5,
					Active:   true,
				}
				break
			}
		}
	}
}

// Update updates all particles
func (ps *ParticleSystem) Update(deltaTime float32) {
	gravity := float32(-9.8)

	for i := range ps.particles {
		if !ps.particles[i].Active {
			continue
		}

		p := &ps.particles[i]

		// Update position
		p.Position = types.NewVec3(
			p.Position.X+p.Velocity.X*deltaTime,
			p.Position.Y+p.Velocity.Y*deltaTime,
			p.Position.Z+p.Velocity.Z*deltaTime,
		)

		// Apply gravity
		p.Velocity.Y = p.Velocity.Y + gravity*deltaTime

		// Update life
		p.Life -= deltaTime
		if p.Life <= 0 {
			p.Active = false
		}
	}
}

// GetActiveParticles returns all active particles for rendering
func (ps *ParticleSystem) GetActiveParticles() []Particle {
	active := make([]Particle, 0)
	for i := range ps.particles {
		if ps.particles[i].Active {
			active = append(active, ps.particles[i])
		}
	}
	return active
}

// Clear removes all particles
func (ps *ParticleSystem) Clear() {
	for i := range ps.particles {
		ps.particles[i].Active = false
	}
}

// Destroy cleans up resources
func (ps *ParticleSystem) Destroy() {
	// No-op for simplified implementation
}

// BlockBreakEffect is a one-shot effect for block breaking
type BlockBreakEffect struct {
	particles []Particle
	origin    types.Vec3
	color     types.Color
	finished  bool
}

// NewBlockBreakEffect creates a block break effect
func NewBlockBreakEffect(pos types.Vec3, blockColor types.Color) *BlockBreakEffect {
	effect := &BlockBreakEffect{
		particles: make([]Particle, 0, 20),
		origin:    pos,
		color:     blockColor,
		finished:  false,
	}

	// Create particles
	for i := 0; i < 20; i++ {
		angle := rand.Float64() * 2 * math.Pi
		speed := float32(rand.Float32()*3 + 2)
		vx := float32(math.Cos(angle)) * speed
		vz := float32(math.Sin(angle)) * speed
		vy := float32(rand.Float32()*4 + 2)

		particle := Particle{
			Position: pos,
			Velocity: types.NewVec3(vx, vy, vz),
			Color:    blockColor,
			Size:     float32(rand.Float32()*0.15 + 0.08),
			Life:     0.8,
			MaxLife:  0.8,
			Active:   true,
		}
		effect.particles = append(effect.particles, particle)
	}

	return effect
}

// Update updates the effect
func (e *BlockBreakEffect) Update(deltaTime float32) bool {
	if e.finished {
		return false
	}

	gravity := float32(-9.8)
	activeCount := 0

	for i := range e.particles {
		if !e.particles[i].Active {
			continue
		}

		p := &e.particles[i]

		// Update position
		p.Position = types.NewVec3(
			p.Position.X+p.Velocity.X*deltaTime,
			p.Position.Y+p.Velocity.Y*deltaTime,
			p.Position.Z+p.Velocity.Z*deltaTime,
		)

		// Apply gravity
		p.Velocity.Y = p.Velocity.Y + gravity*deltaTime

		// Bounce on ground
		if p.Position.Y < e.origin.Y {
			p.Position.Y = e.origin.Y
			p.Velocity.Y = -p.Velocity.Y * 0.5
			p.Velocity.Z = p.Velocity.Z * 0.8
		}

		// Update life
		p.Life -= deltaTime
		if p.Life <= 0 {
			p.Active = false
		} else {
			activeCount++
		}
	}

	if activeCount == 0 {
		e.finished = true
	}

	return !e.finished
}

// GetParticles returns all particles for rendering
func (e *BlockBreakEffect) GetParticles() []Particle {
	return e.particles
}

// IsFinished returns true if effect is done
func (e *BlockBreakEffect) IsFinished() bool {
	return e.finished
}

// EffectManager manages all active effects
type EffectManager struct {
	effects []Effect
}

// Effect interface for all effects
type Effect interface {
	Update(deltaTime float32) bool
	IsFinished() bool
}

// NewEffectManager creates a new effect manager
func NewEffectManager() *EffectManager {
	return &EffectManager{
		effects: make([]Effect, 0),
	}
}

// AddEffect adds an effect to the manager
func (em *EffectManager) AddEffect(effect Effect) {
	em.effects = append(em.effects, effect)
}

// SpawnBlockBreak creates and adds a block break effect
func (em *EffectManager) SpawnBlockBreak(pos types.Vec3, color types.Color) {
	effect := NewBlockBreakEffect(pos, color)
	em.AddEffect(effect)
}

// Update updates all effects and removes finished ones
func (em *EffectManager) Update(deltaTime float32) {
	activeEffects := make([]Effect, 0, len(em.effects))

	for _, effect := range em.effects {
		if effect.Update(deltaTime) {
			activeEffects = append(activeEffects, effect)
		}
	}

	em.effects = activeEffects
}

// Clear removes all effects
func (em *EffectManager) Clear() {
	em.effects = make([]Effect, 0)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
