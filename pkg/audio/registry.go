package audio

import (
	"os"
	"path/filepath"
	"sync"
)

// SoundRegistry manages all loaded sounds
type SoundRegistry struct {
	mu     sync.RWMutex
	sounds map[SoundID]*Sound
}

// NewSoundRegistry creates a new sound registry
func NewSoundRegistry() *SoundRegistry {
	return &SoundRegistry{
		sounds: make(map[SoundID]*Sound),
	}
}

// Register adds a sound to the registry
func (r *SoundRegistry) Register(id SoundID, sound *Sound) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sound.ID = id
	r.sounds[id] = sound
}

// Get retrieves a sound from the registry
func (r *SoundRegistry) Get(id SoundID) *Sound {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.sounds[id]
}

// Unload removes a sound from the registry
func (r *SoundRegistry) Unload(id SoundID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sounds, id)
}

// Clear removes all sounds from the registry
func (r *SoundRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sounds = make(map[SoundID]*Sound)
}

// LoadSound loads a sound file from disk
func LoadSound(path string, volume float32) (*Sound, error) {
	name := filepath.Base(path)

	// Read file data
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return &Sound{
		Name:   name,
		Data:   data,
		Volume: volume,
	}, nil
}
