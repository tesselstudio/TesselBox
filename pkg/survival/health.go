package survival

import (
	"sync"
	"time"
)

// Health represents a player's health system
type Health struct {
	mu           sync.RWMutex
	current      float32
	max          float32
	regeneration float32 // Health per second
	lastDamage   time.Time
	lastHeal     time.Time
	invulnerable bool
	invulnTime   time.Duration
}

// NewHealth creates a new health system
func NewHealth(max float32) *Health {
	return &Health{
		current:      max,
		max:          max,
		regeneration: 0.0,
		lastDamage:   time.Now(),
		lastHeal:     time.Now(),
		invulnerable: false,
		invulnTime:   0,
	}
}

// GetCurrentHealth returns the current health
func (h *Health) GetCurrentHealth() float32 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.current
}

// GetMaxHealth returns the maximum health
func (h *Health) GetMaxHealth() float32 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.max
}

// GetHealthPercentage returns health as a percentage (0-1)
func (h *Health) GetHealthPercentage() float32 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.current / h.max
}

// IsDead returns true if health is 0 or below
func (h *Health) IsDead() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.current <= 0
}

// IsFull returns true if health is at maximum
func (h *Health) IsFull() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.current >= h.max
}

// Damage applies damage to the health
func (h *Health) Damage(amount float32, damageType DamageType) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.invulnerable || h.current <= 0 {
		return false
	}

	// Apply damage with type modifiers
	actualDamage := amount * damageType.GetMultiplier()

	h.current -= actualDamage
	if h.current < 0 {
		h.current = 0
	}

	h.lastDamage = time.Now()

	// Set temporary invulnerability
	h.SetInvulnerable(1.0 * time.Second)

	return true
}

// Heal restores health
func (h *Health) Heal(amount float32) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.current >= h.max || h.current <= 0 {
		return false
	}

	h.current += amount
	if h.current > h.max {
		h.current = h.max
	}

	h.lastHeal = time.Now()

	return true
}

// SetMaxHealth sets the maximum health
func (h *Health) SetMaxHealth(max float32) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.max = max
	if h.current > max {
		h.current = max
	}
}

// SetRegeneration sets the health regeneration rate
func (h *Health) SetRegeneration(rate float32) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.regeneration = rate
}

// SetInvulnerable sets temporary invulnerability
func (h *Health) SetInvulnerable(duration time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.invulnerable = true
	h.invulnTime = duration
}

// Update updates the health system (call every frame)
func (h *Health) Update(deltaTime float32) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Update invulnerability
	if h.invulnerable {
		h.invulnTime -= time.Duration(deltaTime) * time.Second
		if h.invulnTime <= 0 {
			h.invulnerable = false
			h.invulnTime = 0
		}
	}

	// Apply regeneration
	if h.regeneration > 0 && h.current > 0 && h.current < h.max {
		h.current += h.regeneration * deltaTime
		if h.current > h.max {
			h.current = h.max
		}
	}
}

// Reset resets health to maximum
func (h *Health) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.current = h.max
	h.lastDamage = time.Now()
	h.lastHeal = time.Now()
	h.invulnerable = false
	h.invulnTime = 0
}

// DamageType represents different types of damage
type DamageType int

const (
	DamageTypePhysical DamageType = iota
	DamageTypeFire
	DamageTypeWater
	DamageTypeFall
	DamageTypePoison
	DamageTypeMagic
	DamageTypeExplosion
	DamageTypeEnvironmental
)

// GetMultiplier returns the damage multiplier for this type
func (dt DamageType) GetMultiplier() float32 {
	switch dt {
	case DamageTypePhysical:
		return 1.0
	case DamageTypeFire:
		return 1.0
	case DamageTypeWater:
		return 0.5
	case DamageTypeFall:
		return 1.0
	case DamageTypePoison:
		return 0.8
	case DamageTypeMagic:
		return 1.2
	case DamageTypeExplosion:
		return 1.5
	case DamageTypeEnvironmental:
		return 0.7
	default:
		return 1.0
	}
}

// Hunger represents a player's hunger system
type Hunger struct {
	mu         sync.RWMutex
	current    float32
	max        float32
	saturation float32
	exhaustion float32
	lastFood   time.Time
	starving   bool
}

// NewHunger creates a new hunger system
func NewHunger(max float32) *Hunger {
	return &Hunger{
		current:    max,
		max:        max,
		saturation: 5.0,
		exhaustion: 0.0,
		lastFood:   time.Now(),
		starving:   false,
	}
}

// GetCurrentHunger returns the current hunger level
func (hu *Hunger) GetCurrentHunger() float32 {
	hu.mu.RLock()
	defer hu.mu.RUnlock()
	return hu.current
}

// GetMaxHunger returns the maximum hunger
func (hu *Hunger) GetMaxHunger() float32 {
	hu.mu.RLock()
	defer hu.mu.RUnlock()
	return hu.max
}

// GetSaturation returns the current saturation level
func (hu *Hunger) GetSaturation() float32 {
	hu.mu.RLock()
	defer hu.mu.RUnlock()
	return hu.saturation
}

// GetHungerPercentage returns hunger as a percentage (0-1)
func (hu *Hunger) GetHungerPercentage() float32 {
	hu.mu.RLock()
	defer hu.mu.RUnlock()
	return hu.current / hu.max
}

// IsStarving returns true if the player is starving
func (hu *Hunger) IsStarving() bool {
	hu.mu.RLock()
	defer hu.mu.RUnlock()
	return hu.starving
}

// IsFull returns true if hunger is at maximum
func (hu *Hunger) IsFull() bool {
	hu.mu.RLock()
	defer hu.mu.RUnlock()
	return hu.current >= hu.max
}

// Eat consumes food and restores hunger
func (hu *Hunger) Eat(hungerRestore, saturationRestore float32) bool {
	hu.mu.Lock()
	defer hu.mu.Unlock()

	if hu.current >= hu.max {
		return false
	}

	hu.current += hungerRestore
	if hu.current > hu.max {
		hu.current = hu.max
	}

	hu.saturation += saturationRestore
	if hu.saturation > hu.current {
		hu.saturation = hu.current
	}

	hu.lastFood = time.Now()
	hu.starving = false

	return true
}

// Exhaust increases exhaustion (from activities)
func (hu *Hunger) Exhaust(amount float32) {
	hu.mu.Lock()
	defer hu.mu.Unlock()

	hu.exhaustion += amount

	// When exhaustion reaches 4.0, reduce saturation or hunger
	if hu.exhaustion >= 4.0 {
		if hu.saturation > 0 {
			hu.saturation -= 1.0
			if hu.saturation < 0 {
				hu.saturation = 0
			}
		} else {
			hu.current -= 1.0
			if hu.current < 0 {
				hu.current = 0
			}
		}

		hu.exhaustion -= 4.0
	}

	// Check for starvation
	if hu.current <= 0 {
		hu.starving = true
	}
}

// Update updates the hunger system (call every frame)
func (hu *Hunger) Update(deltaTime float32) {
	hu.mu.Lock()
	defer hu.mu.Unlock()

	// Natural hunger decrease over time
	if hu.current > 0 {
		hu.current -= 0.017 * deltaTime // 1 hunger point per ~60 seconds
		if hu.current < 0 {
			hu.current = 0
		}
	}

	// Saturation decreases first, then hunger
	if hu.saturation > 0 {
		hu.saturation -= 0.1 * deltaTime
		if hu.saturation < 0 {
			hu.saturation = 0
		}
	}

	// Check for starvation
	if hu.current <= 0 && !hu.starving {
		hu.starving = true
	}
}

// Reset resets hunger to maximum
func (hu *Hunger) Reset() {
	hu.mu.Lock()
	defer hu.mu.Unlock()

	hu.current = hu.max
	hu.saturation = 5.0
	hu.exhaustion = 0.0
	hu.lastFood = time.Now()
	hu.starving = false
}

// Stamina represents a player's stamina system
type Stamina struct {
	mu           sync.RWMutex
	current      float32
	max          float32
	regeneration float32
	exhausted    bool
	lastUse      time.Time
}

// NewStamina creates a new stamina system
func NewStamina(max float32) *Stamina {
	return &Stamina{
		current:      max,
		max:          max,
		regeneration: 10.0, // 10 stamina per second
		exhausted:    false,
		lastUse:      time.Now(),
	}
}

// GetCurrentStamina returns the current stamina
func (s *Stamina) GetCurrentStamina() float32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// GetMaxStamina returns the maximum stamina
func (s *Stamina) GetMaxStamina() float32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.max
}

// GetStaminaPercentage returns stamina as a percentage (0-1)
func (s *Stamina) GetStaminaPercentage() float32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current / s.max
}

// IsExhausted returns true if stamina is exhausted
func (s *Stamina) IsExhausted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.exhausted
}

// IsFull returns true if stamina is at maximum
func (s *Stamina) IsFull() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current >= s.max
}

// Use consumes stamina for an action
func (s *Stamina) Use(amount float32) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.exhausted || s.current <= 0 {
		return false
	}

	s.current -= amount
	if s.current < 0 {
		s.current = 0
		s.exhausted = true
	}

	s.lastUse = time.Now()

	return true
}

// Update updates the stamina system (call every frame)
func (s *Stamina) Update(deltaTime float32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Regenerate stamina if not exhausted or if recovering from exhaustion
	if s.current < s.max {
		regenRate := s.regeneration

		// Slower regeneration when exhausted
		if s.exhausted {
			regenRate *= 0.5
		}

		s.current += regenRate * deltaTime
		if s.current > s.max {
			s.current = s.max
		}

		// Recover from exhaustion at 50% stamina
		if s.exhausted && s.current >= s.max*0.5 {
			s.exhausted = false
		}
	}
}

// Reset resets stamina to maximum
func (s *Stamina) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.current = s.max
	s.exhausted = false
	s.lastUse = time.Now()
}

// PlayerStats represents all player survival stats
type PlayerStats struct {
	Health     *Health
	Hunger     *Hunger
	Stamina    *Stamina
	Experience *Experience
}

// NewPlayerStats creates a new player stats system
func NewPlayerStats(maxHealth, maxHunger, maxStamina float32) *PlayerStats {
	return &PlayerStats{
		Health:     NewHealth(maxHealth),
		Hunger:     NewHunger(maxHunger),
		Stamina:    NewStamina(maxStamina),
		Experience: NewExperience(),
	}
}

// Update updates all player stats (call every frame)
func (ps *PlayerStats) Update(deltaTime float32) {
	ps.Health.Update(deltaTime)
	ps.Hunger.Update(deltaTime)
	ps.Stamina.Update(deltaTime)
	ps.Experience.Update(deltaTime)

	// Apply hunger damage if starving
	if ps.Hunger.IsStarving() {
		ps.Health.Damage(1.0*deltaTime, DamageTypeEnvironmental)
	}
}

// Reset resets all player stats
func (ps *PlayerStats) Reset() {
	ps.Health.Reset()
	ps.Hunger.Reset()
	ps.Stamina.Reset()
	ps.Experience.Reset()
}

// Experience represents player experience and leveling
type Experience struct {
	mu               sync.RWMutex
	current          int32
	level            int32
	experienceToNext int32
	total            int64
	skillPoints      int32
}

// NewExperience creates a new experience system
func NewExperience() *Experience {
	return &Experience{
		current:          0,
		level:            1,
		experienceToNext: 100,
		total:            0,
		skillPoints:      0,
	}
}

// GetCurrentExperience returns current experience
func (e *Experience) GetCurrentExperience() int32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.current
}

// GetLevel returns current level
func (e *Experience) GetLevel() int32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.level
}

// GetExperienceToNext returns experience needed for next level
func (e *Experience) GetExperienceToNext() int32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.experienceToNext
}

// GetTotalExperience returns total experience earned
func (e *Experience) GetTotalExperience() int64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.total
}

// GetSkillPoints returns available skill points
func (e *Experience) GetSkillPoints() int32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.skillPoints
}

// AddExperience adds experience points
func (e *Experience) AddExperience(amount int32) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.current += amount
	e.total += int64(amount)

	leveledUp := false
	for e.current >= e.experienceToNext {
		e.current -= e.experienceToNext
		e.level++
		e.skillPoints++
		e.experienceToNext = calculateExperienceForLevel(e.level)
		leveledUp = true
	}

	return leveledUp
}

// SpendSkillPoint spends a skill point
func (e *Experience) SpendSkillPoint() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.skillPoints <= 0 {
		return false
	}

	e.skillPoints--
	return true
}

// Update updates the experience system
func (e *Experience) Update(deltaTime float32) {
	// Experience doesn't need per-frame updates
}

// Reset resets experience system
func (e *Experience) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.current = 0
	e.level = 1
	e.experienceToNext = 100
	e.total = 0
	e.skillPoints = 0
}

// calculateExperienceForLevel calculates experience needed for a level
func calculateExperienceForLevel(level int32) int32 {
	// Exponential growth: 100 * 1.5^(level-1)
	base := int32(100)
	for i := int32(1); i < level; i++ {
		base = base * 3 / 2
	}
	return base
}
