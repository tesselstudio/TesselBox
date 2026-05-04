package effects

import (
	"math"
	"math/rand"
	"time"

	"kaijuengine.com/engine"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

// Particle represents a single particle
type Particle struct {
	Position  matrix.Vec3
	Velocity  matrix.Vec3
	Color     matrix.Color
	Size      float32
	Life      float32
	MaxLife   float32
	Active    bool
}

// ParticleSystem manages a collection of particles
type ParticleSystem struct {
	particles []Particle
	host      *engine.Host
	drawing   *rendering.Drawing
	entity    *engine.Entity
}

// NewParticleSystem creates a new particle system
func NewParticleSystem(host *engine.Host, maxParticles int) *ParticleSystem {
	ps := &ParticleSystem{
		particles: make([]Particle, maxParticles),
		host:      host,
	}

	// Create entity for the particle system
	ps.entity = engine.NewEntity(host.WorkGroup())

	return ps
}

// SpawnBlockBreakParticles creates particles when a block is broken
func (ps *ParticleSystem) SpawnBlockBreakParticles(pos matrix.Vec3, blockColor matrix.Color, count int) {
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
					Velocity: matrix.NewVec3(vx, vy, vz),
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
func (ps *ParticleSystem) SpawnFootstepParticles(pos matrix.Vec3, count int) {
	for i := 0; i < count; i++ {
		for j := range ps.particles {
			if !ps.particles[j].Active {
				// Random velocity upward
				vx := float32(rand.Float32()*0.5 - 0.25)
				vz := float32(rand.Float32()*0.5 - 0.25)
				vy := float32(rand.Float32() * 0.5)

				ps.particles[j] = Particle{
					Position: pos,
					Velocity: matrix.NewVec3(vx, vy, vz),
					Color:    matrix.NewColor(0.7, 0.7, 0.6, 0.8),
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
		p.Position = matrix.NewVec3(
			p.Position.X()+p.Velocity.X()*deltaTime,
			p.Position.Y()+p.Velocity.Y()*deltaTime,
			p.Position.Z()+p.Velocity.Z()*deltaTime,
		)

		// Apply gravity
		p.Velocity.SetY(p.Velocity.Y() + gravity*deltaTime)

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
	if ps.entity != nil {
		ps.host.DestroyEntity(ps.entity)
	}
}

// BlockBreakEffect is a one-shot effect for block breaking
type BlockBreakEffect struct {
	particles []Particle
	origin    matrix.Vec3
	color     matrix.Color
	finished  bool
}

// NewBlockBreakEffect creates a block break effect
func NewBlockBreakEffect(pos matrix.Vec3, blockColor matrix.Color) *BlockBreakEffect {
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
			Velocity: matrix.NewVec3(vx, vy, vz),
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
		p.Position = matrix.NewVec3(
			p.Position.X()+p.Velocity.X()*deltaTime,
			p.Position.Y()+p.Velocity.Y()*deltaTime,
			p.Position.Z()+p.Velocity.Z()*deltaTime,
		)

		// Apply gravity
		p.Velocity.SetY(p.Velocity.Y() + gravity*deltaTime)

		// Bounce on ground
		if p.Position.Y() < e.origin.Y() {
			p.Position.SetY(e.origin.Y())
			p.Velocity.SetY(-p.Velocity.Y() * 0.5)
			p.Velocity.SetX(p.Velocity.X() * 0.8)
			p.Velocity.SetZ(p.Velocity.Z() * 0.8)
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
	host    *engine.Host
}

// Effect interface for all effects
type Effect interface {
	Update(deltaTime float32) bool
	IsFinished() bool
}

// NewEffectManager creates a new effect manager
func NewEffectManager(host *engine.Host) *EffectManager {
	return &EffectManager{
		effects: make([]Effect, 0),
		host:    host,
	}
}

// AddEffect adds an effect to the manager
func (em *EffectManager) AddEffect(effect Effect) {
	em.effects = append(em.effects, effect)
}

// SpawnBlockBreak creates and adds a block break effect
func (em *EffectManager) SpawnBlockBreak(pos matrix.Vec3, color matrix.Color) {
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
