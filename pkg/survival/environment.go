package survival

import (
	"math"
	"sync"
	"time"
)

// Environment represents the environmental conditions
type Environment struct {
	mu          sync.RWMutex
	timeOfDay   float32 // 0-24 hours
	day         int32
	temperature float32 // Celsius
	humidity    float32 // 0-100%
	weather     WeatherType
	season      Season
	biome       BiomeType
}

// NewEnvironment creates a new environment system
func NewEnvironment() *Environment {
	return &Environment{
		timeOfDay:   6.0, // Start at 6 AM
		day:         0,
		temperature: 20.0,
		humidity:    50.0,
		weather:     WeatherClear,
		season:      SeasonSpring,
		biome:       BiomePlains,
	}
}

// GetTimeOfDay returns the current time of day (0-24)
func (e *Environment) GetTimeOfDay() float32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.timeOfDay
}

// GetDay returns the current day number
func (e *Environment) GetDay() int32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.day
}

// GetTemperature returns the current temperature in Celsius
func (e *Environment) GetTemperature() float32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.temperature
}

// GetHumidity returns the current humidity percentage
func (e *Environment) GetHumidity() float32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.humidity
}

// GetWeather returns the current weather
func (e *Environment) GetWeather() WeatherType {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.weather
}

// GetSeason returns the current season
func (e *Environment) GetSeason() Season {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.season
}

// GetBiome returns the current biome
func (e *Environment) GetBiome() BiomeType {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.biome
}

// IsDaytime returns true if it's daytime
func (e *Environment) IsDaytime() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.timeOfDay >= 6.0 && e.timeOfDay < 18.0
}

// IsNighttime returns true if it's nighttime
func (e *Environment) IsNighttime() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.timeOfDay >= 18.0 || e.timeOfDay < 6.0
}

// GetSunIntensity returns sun intensity (0-1)
func (e *Environment) GetSunIntensity() float32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if !e.IsDaytime() {
		return 0.0
	}
	
	// Peak sun at noon (12:00)
	normalized := math.Abs(float64(e.timeOfDay - 12.0)) / 6.0
	return float32(1.0 - normalized)
}

// Update updates the environment (call every frame)
func (e *Environment) Update(deltaTime float32) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// Update time of day (1 real second = 1 game minute)
	e.timeOfDay += deltaTime / 60.0
	
	if e.timeOfDay >= 24.0 {
		e.timeOfDay -= 24.0
		e.day++
		
		// Update season every 30 days
		if e.day%30 == 0 {
			e.season = Season((int(e.season) + 1) % 4)
		}
	}
	
	// Update temperature based on time and season
	e.updateTemperature()
	
	// Update weather occasionally
	if time.Now().Unix()%300 == 0 { // Every 5 minutes
		e.updateWeather()
	}
}

// updateTemperature updates temperature based on time and season
func (e *Environment) updateTemperature() {
	baseTemp := e.getBaseTemperature()
	
	// Time-based variation
	timeVariation := float32(math.Cos(float64(e.timeOfDay-14.0) * math.Pi / 12.0)) * 8.0
	
	// Weather variation
	weatherVariation := e.weather.GetTemperatureModifier()
	
	// Biome variation
	biomeVariation := e.biome.GetTemperatureModifier()
	
	e.temperature = baseTemp + timeVariation + weatherVariation + biomeVariation
}

// updateWeather updates weather conditions
func (e *Environment) updateWeather() {
	// Simple weather logic - can be made more complex
	if e.humidity > 70 && e.weather == WeatherClear {
		e.weather = WeatherCloudy
	} else if e.humidity > 85 && e.weather == WeatherCloudy {
		e.weather = WeatherRain
	} else if e.humidity < 30 && e.weather == WeatherCloudy {
		e.weather = WeatherClear
	}
}

// getBaseTemperature returns base temperature for current season
func (e *Environment) getBaseTemperature() float32 {
	switch e.season {
	case SeasonSpring:
		return 15.0
	case SeasonSummer:
		return 25.0
	case SeasonAutumn:
		return 10.0
	case SeasonWinter:
		return 0.0
	default:
		return 20.0
	}
}

// SetBiome sets the current biome
func (e *Environment) SetBiome(biome BiomeType) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.biome = biome
}

// WeatherType represents different weather conditions
type WeatherType int

const (
	WeatherClear WeatherType = iota
	WeatherCloudy
	WeatherRain
	WeatherStorm
	WeatherSnow
	WeatherFog
)

// GetTemperatureModifier returns temperature modifier for weather
func (wt WeatherType) GetTemperatureModifier() float32 {
	switch wt {
	case WeatherClear:
		return 0.0
	case WeatherCloudy:
		return -2.0
	case WeatherRain:
		return -5.0
	case WeatherStorm:
		return -8.0
	case WeatherSnow:
		return -10.0
	case WeatherFog:
		return -1.0
	default:
		return 0.0
	}
}

// GetVisibilityModifier returns visibility modifier for weather
func (wt WeatherType) GetVisibilityModifier() float32 {
	switch wt {
	case WeatherClear:
		return 1.0
	case WeatherCloudy:
		return 0.8
	case WeatherRain:
		return 0.6
	case WeatherStorm:
		return 0.4
	case WeatherSnow:
		return 0.7
	case WeatherFog:
		return 0.3
	default:
		return 1.0
	}
}

// Season represents different seasons
type Season int

const (
	SeasonSpring Season = iota
	SeasonSummer
	SeasonAutumn
	SeasonWinter
)

// BiomeType represents different biomes
type BiomeType int

const (
	BiomePlains BiomeType = iota
	BiomeForest
	BiomeDesert
	BiomeTundra
	BiomeMountains
	BiomeOcean
	BiomeSwamp
	BiomeJungle
)

// GetTemperatureModifier returns temperature modifier for biome
func (bt BiomeType) GetTemperatureModifier() float32 {
	switch bt {
	case BiomePlains:
		return 0.0
	case BiomeForest:
		return -2.0
	case BiomeDesert:
		return 10.0
	case BiomeTundra:
		return -15.0
	case BiomeMountains:
		return -8.0
	case BiomeOcean:
		return 0.0
	case BiomeSwamp:
		return 2.0
	case BiomeJungle:
		return 5.0
	default:
		return 0.0
	}
}

// EnvironmentalHazard represents environmental hazards
type EnvironmentalHazard struct {
	Type        HazardType
	Position    HazardPosition
	Radius      float32
	Damage      float32
	Effect      HazardEffect
	Duration    float32
	Active      bool
	StartTime   time.Time
}

// HazardType represents different hazard types
type HazardType int

const (
	HazardTypeFire HazardType = iota
	HazardTypeWater
	HazardTypeLava
	HazardTypePoison
	HazardTypeRadiation
	HazardTypeCold
	HazardTypeHeat
)

// HazardPosition represents hazard position
type HazardPosition struct {
	X, Y, Z float32
}

// HazardEffect represents the effect of a hazard
type HazardEffect int

const (
	HazardEffectDamage HazardEffect = iota
	HazardEffectSlow
	HazardEffectPoison
	HazardEffectBurn
	HazardEffectFreeze
)

// HazardManager manages environmental hazards
type HazardManager struct {
	mu      sync.RWMutex
	hazards map[int32]*EnvironmentalHazard
	nextID  int32
}

// NewHazardManager creates a new hazard manager
func NewHazardManager() *HazardManager {
	return &HazardManager{
		hazards: make(map[int32]*EnvironmentalHazard),
		nextID:  1,
	}
}

// AddHazard adds a new environmental hazard
func (hm *HazardManager) AddHazard(hazardType HazardType, position HazardPosition, radius, damage, duration float32, effect HazardEffect) int32 {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	
	hazard := &EnvironmentalHazard{
		Type:      hazardType,
		Position:  position,
		Radius:    radius,
		Damage:    damage,
		Effect:    effect,
		Duration:  duration,
		Active:    true,
		StartTime: time.Now(),
	}
	
	id := hm.nextID
	hm.nextID++
	
	hm.hazards[id] = hazard
	
	return id
}

// RemoveHazard removes a hazard
func (hm *HazardManager) RemoveHazard(id int32) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	
	delete(hm.hazards, id)
}

// GetHazardsInArea returns hazards affecting a position
func (hm *HazardManager) GetHazardsInArea(position HazardPosition, radius float32) []*EnvironmentalHazard {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	
	var affectingHazards []*EnvironmentalHazard
	
	for _, hazard := range hm.hazards {
		if !hazard.Active {
			continue
		}
		
		distance := hm.calculateDistance(position, hazard.Position)
		if distance <= hazard.Position.Z+radius {
			affectingHazards = append(affectingHazards, hazard)
		}
	}
	
	return affectingHazards
}

// Update updates all hazards (call every frame)
func (hm *HazardManager) Update(deltaTime float32) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	
	now := time.Now()
	
	for id, hazard := range hm.hazards {
		if !hazard.Active {
			continue
		}
		
		// Check if hazard has expired
		elapsed := float32(now.Sub(hazard.StartTime).Seconds())
		if elapsed >= hazard.Duration {
			hazard.Active = false
			delete(hm.hazards, id)
		}
	}
}

// calculateDistance calculates distance between two positions
func (hm *HazardManager) calculateDistance(pos1, pos2 HazardPosition) float32 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	dz := pos1.Z - pos2.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

// TemperatureSystem manages player temperature
type TemperatureSystem struct {
	mu           sync.RWMutex
	current      float32
	target       float32
	comfort      float32
	resistance   float32
	lastUpdate   time.Time
}

// NewTemperatureSystem creates a new temperature system
func NewTemperatureSystem() *TemperatureSystem {
	return &TemperatureSystem{
		current:    37.0, // Normal body temperature
		target:     37.0,
		comfort:    20.0, // Comfortable ambient temperature
		resistance: 1.0,
		lastUpdate: time.Now(),
	}
}

// GetCurrentTemperature returns current body temperature
func (ts *TemperatureSystem) GetCurrentTemperature() float32 {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.current
}

// SetAmbientTemperature sets the target temperature based on ambient conditions
func (ts *TemperatureSystem) SetAmbientTemperature(ambient float32) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	// Calculate target body temperature based on ambient
	// This is simplified - in reality would consider clothing, activity, etc.
	ts.target = 37.0 + (ambient - ts.comfort) * 0.1
	ts.target = float32(math.Max(float64(35.0), math.Min(float64(40.0), float64(ts.target))))
}

// Update updates the temperature system (call every frame)
func (ts *TemperatureSystem) Update(deltaTime float32) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	// Gradually adjust current temperature towards target
	diff := ts.target - ts.current
	ts.current += diff * deltaTime * 0.1 * ts.resistance
}

// IsTooHot returns true if temperature is dangerously high
func (ts *TemperatureSystem) IsTooHot() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.current > 39.0
}

// IsTooCold returns true if temperature is dangerously low
func (ts *TemperatureSystem) IsTooCold() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.current < 35.0
}

// IsComfortable returns true if temperature is in comfortable range
func (ts *TemperatureSystem) IsComfortable() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.current >= 36.0 && ts.current <= 38.0
}
