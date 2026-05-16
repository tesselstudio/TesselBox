package audio

import (
	"fmt"
	"sync"
)

// SFXEngine handles sound effect playback using a simple Go-based audio engine
// This is a simplified implementation that uses a stub backend
// In production, this would integrate with oto, beep, or Kaiju's audio system
type SFXEngine struct {
	mu       sync.RWMutex
	loaded   bool
	volume   float32
	sources  map[string]*AudioSource
}

// AudioSource represents a loaded audio source
type AudioSource struct {
	Data   []byte
	Volume float32
	Format AudioFormat
}

// AudioFormat represents the audio file format
type AudioFormat int

const (
	FormatUnknown AudioFormat = iota
	FormatWAV
	FormatOGG
	FormatMP3
)

// NewSFXEngine creates a new sound effects engine
func NewSFXEngine() *SFXEngine {
	return &SFXEngine{
		sources: make(map[string]*AudioSource),
		volume:  1.0,
	}
}

// Initialize sets up the audio engine
func (e *SFXEngine) Initialize() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.loaded {
		return nil
	}

	// In a full implementation, this would:
	// 1. Initialize the audio device
	// 2. Set up audio buffers
	// 3. Configure audio context

	e.loaded = true
	return nil
}

// Shutdown cleans up audio resources
func (e *SFXEngine) Shutdown() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.loaded {
		return
	}

	// Clean up all sources
	for k := range e.sources {
		delete(e.sources, k)
	}

	e.loaded = false
}

// Play plays a sound at the specified volume
func (e *SFXEngine) Play(sound *Sound, volume float32) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.loaded {
		return
	}

	// In a full implementation, this would:
	// 1. Find or create an audio buffer for the sound
	// 2. Play the sound at the specified volume
	// 3. Manage concurrent playback

	finalVolume := e.volume * volume * sound.Volume
	if finalVolume <= 0 {
		return
	}

	// Stub: just log for now
	_ = fmt.Sprintf("Playing sound %s at volume %.2f", sound.Name, finalVolume)
}

// PlayWithPan plays a sound with stereo panning
func (e *SFXEngine) PlayWithPan(sound *Sound, volume, pan float32) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.loaded {
		return
	}

	finalVolume := e.volume * volume * sound.Volume
	if finalVolume <= 0 {
		return
	}

	// In a full implementation, this would apply stereo panning
	// pan: -1.0 (full left) to 1.0 (full right), 0.0 = center

	_ = fmt.Sprintf("Playing sound %s at volume %.2f, pan %.2f", sound.Name, finalVolume, pan)
}

// LoadSource loads an audio source from file data
func (e *SFXEngine) LoadSource(name string, data []byte, format AudioFormat) (*AudioSource, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	source := &AudioSource{
		Data:   data,
		Volume: 1.0,
		Format: format,
	}

	e.sources[name] = source
	return source, nil
}

// UnloadSource removes an audio source
func (e *SFXEngine) UnloadSource(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.sources, name)
}

// SetVolume sets the engine volume
func (e *SFXEngine) SetVolume(volume float32) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.volume = clamp(volume, 0.0, 1.0)
}

// StopAll stops all playing sounds
func (e *SFXEngine) StopAll() {
	e.mu.Lock()
	defer e.mu.Unlock()

	// In a full implementation, stop all audio sources
}
