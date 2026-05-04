package audio

import (
	"fmt"
	"path/filepath"
	"sync"
)

// Manager handles all audio playback and resource management
type Manager struct {
	mu sync.RWMutex

	// Audio engines
	sfxEngine   *SFKEngine
	musicEngine *MusicEngine

	// Settings
	masterVolume float32
	sfxVolume    float32
	musicVolume  float32
	muted        bool

	// Asset paths
	basePath string

	// Registry
	soundRegistry *SoundRegistry

	// Runtime state
	initialized bool
}

var (
	instance *Manager
	once     sync.Once
)

// GetManager returns the singleton audio manager instance
func GetManager() *Manager {
	once.Do(func() {
		instance = &Manager{
			basePath:     "game_content/audio",
			masterVolume: 1.0,
			sfxVolume:    1.0,
			musicVolume:  0.7,
			soundRegistry: NewSoundRegistry(),
		}
	})
	return instance
}

// Initialize sets up the audio system
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return nil
	}

	// Initialize SFX engine (using beep for now)
	m.sfxEngine = NewSFKEngine()
	if err := m.sfxEngine.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize SFX engine: %w", err)
	}

	// Initialize music engine
	m.musicEngine = NewMusicEngine(m.basePath)
	if err := m.musicEngine.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize music engine: %w", err)
	}

	// Load default sounds
	if err := m.loadDefaultSounds(); err != nil {
		return fmt.Errorf("failed to load default sounds: %w", err)
	}

	m.initialized = true
	return nil
}

// Shutdown cleans up audio resources
func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return
	}

	if m.sfxEngine != nil {
		m.sfxEngine.Shutdown()
	}

	if m.musicEngine != nil {
		m.musicEngine.Shutdown()
	}

	m.initialized = false
}

// PlaySFX plays a sound effect
func (m *Manager) PlaySFX(soundID SoundID) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized || m.muted {
		return
	}

	sound := m.soundRegistry.Get(soundID)
	if sound == nil {
		return
	}

	volume := m.masterVolume * m.sfxVolume * sound.Volume
	m.sfxEngine.Play(sound, volume)
}

// PlaySFXAt plays a sound effect at a 3D position
func (m *Manager) PlaySFXAt(soundID SoundID, listenerPos, soundPos Vec3) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized || m.muted {
		return
	}

	sound := m.soundRegistry.Get(soundID)
	if sound == nil {
		return
	}

	// Calculate positional audio
	attenuation := CalculateAttenuation(listenerPos, soundPos)
	pan := CalculatePan(listenerPos, soundPos)

	volume := m.masterVolume * m.sfxVolume * sound.Volume * attenuation
	m.sfxEngine.PlayWithPan(sound, volume, pan)
}

// PlayMusic starts playing background music
func (m *Manager) PlayMusic(trackName string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized || m.muted {
		return
	}

	volume := m.masterVolume * m.musicVolume
	m.musicEngine.Play(trackName, volume)
}

// StopMusic stops the current music
func (m *Manager) StopMusic() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return
	}

	m.musicEngine.Stop()
}

// PauseMusic pauses the music
func (m *Manager) PauseMusic() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return
	}

	m.musicEngine.Pause()
}

// ResumeMusic resumes paused music
func (m *Manager) ResumeMusic() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return
	}

	m.musicEngine.Resume()
}

// SetMasterVolume sets the master volume (0.0 - 1.0)
func (m *Manager) SetMasterVolume(volume float32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.masterVolume = clamp(volume, 0.0, 1.0)
	m.updateVolumes()
}

// SetSFXVolume sets the sound effects volume (0.0 - 1.0)
func (m *Manager) SetSFXVolume(volume float32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sfxVolume = clamp(volume, 0.0, 1.0)
}

// SetMusicVolume sets the music volume (0.0 - 1.0)
func (m *Manager) SetMusicVolume(volume float32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.musicVolume = clamp(volume, 0.0, 1.0)
	if m.musicEngine != nil {
		m.musicEngine.SetVolume(m.masterVolume * m.musicVolume)
	}
}

// SetMuted sets the global mute state
func (m *Manager) SetMuted(muted bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.muted = muted
	if muted {
		if m.musicEngine != nil {
			m.musicEngine.SetVolume(0)
		}
	} else {
		m.updateVolumes()
	}
}

// IsMuted returns the current mute state
func (m *Manager) IsMuted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.muted
}

// GetMasterVolume returns the master volume
func (m *Manager) GetMasterVolume() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.masterVolume
}

// GetSFXVolume returns the SFX volume
func (m *Manager) GetSFXVolume() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sfxVolume
}

// GetMusicVolume returns the music volume
func (m *Manager) GetMusicVolume() float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.musicVolume
}

// TriggerGameEvent plays sounds for game events
func (m *Manager) TriggerGameEvent(event GameEvent) {
	switch event {
	case EventBlockBreak:
		m.PlaySFX(SFXBlockBreak)
	case EventBlockPlace:
		m.PlaySFX(SFXBlockPlace)
	case EventFootstep:
		m.PlaySFX(SFXFootstep)
	case EventUIClick:
		m.PlaySFX(SFXUIClick)
	case EventUIHover:
		m.PlaySFX(SFXUIHover)
	case EventItemPickup:
		m.PlaySFX(SFXItemPickup)
	case EventPlayerDamage:
		m.PlaySFX(SFXPlayerDamage)
	case EventJump:
		m.PlaySFX(SFXJump)
	}
}

// Update should be called each frame to update streaming audio
func (m *Manager) Update() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return
	}

	m.musicEngine.Update()
}

// Private methods

func (m *Manager) loadDefaultSounds() error {
	// Define default sounds with their paths
	defaults := []struct {
		id       SoundID
		filename string
		volume   float32
	}{
		{SFXBlockBreak, "sfx/block_break.ogg", 1.0},
		{SFXBlockPlace, "sfx/block_place.ogg", 1.0},
		{SFXFootstep, "sfx/footstep.ogg", 0.6},
		{SFXUIClick, "sfx/ui_click.ogg", 0.8},
		{SFXUIHover, "sfx/ui_hover.ogg", 0.4},
		{SFXItemPickup, "sfx/item_pickup.ogg", 0.7},
		{SFXPlayerDamage, "sfx/player_damage.ogg", 1.0},
		{SFXJump, "sfx/jump.ogg", 0.5},
	}

	for _, def := range defaults {
		path := filepath.Join(m.basePath, def.filename)
		sound, err := LoadSound(path, def.volume)
		if err != nil {
			// Log but don't fail - sounds are optional
			fmt.Printf("Warning: could not load sound %s: %v\n", def.filename, err)
			continue
		}
		m.soundRegistry.Register(def.id, sound)
	}

	return nil
}

func (m *Manager) updateVolumes() {
	if m.musicEngine != nil {
		m.musicEngine.SetVolume(m.masterVolume * m.musicVolume)
	}
}

func clamp(value, min, max float32) float32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
