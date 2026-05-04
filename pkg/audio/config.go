package audio

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Config holds audio settings that can be saved/loaded
type Config struct {
	mu sync.RWMutex

	MasterVolume float32 `json:"master_volume"`
	SFXVolume    float32 `json:"sfx_volume"`
	MusicVolume  float32 `json:"music_volume"`
	Muted        bool    `json:"muted"`

	// Audio quality settings
	Enable3DAudio bool `json:"enable_3d_audio"`
	HighQuality   bool `json:"high_quality"`

	// Device settings
	OutputDevice string `json:"output_device"`
}

// DefaultConfig returns the default audio configuration
func DefaultConfig() *Config {
	return &Config{
		MasterVolume:  1.0,
		SFXVolume:     1.0,
		MusicVolume:   0.7,
		Muted:         false,
		Enable3DAudio: true,
		HighQuality:   true,
		OutputDevice:  "default",
	}
}

// LoadConfig loads audio configuration from file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save saves the audio configuration to file
func (c *Config) Save(path string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Apply applies this configuration to the audio manager
func (c *Config) Apply(manager *Manager) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	manager.SetMasterVolume(c.MasterVolume)
	manager.SetSFXVolume(c.SFXVolume)
	manager.SetMusicVolume(c.MusicVolume)
	manager.SetMuted(c.Muted)
}

// Sync syncs this configuration from the audio manager
func (c *Config) Sync(manager *Manager) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.MasterVolume = manager.GetMasterVolume()
	c.SFXVolume = manager.GetSFXVolume()
	c.MusicVolume = manager.GetMusicVolume()
	c.Muted = manager.IsMuted()
}

// Getters and Setters

func (c *Config) GetMasterVolume() float32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MasterVolume
}

func (c *Config) SetMasterVolume(v float32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.MasterVolume = clamp(v, 0.0, 1.0)
}

func (c *Config) GetSFXVolume() float32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SFXVolume
}

func (c *Config) SetSFXVolume(v float32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SFXVolume = clamp(v, 0.0, 1.0)
}

func (c *Config) GetMusicVolume() float32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MusicVolume
}

func (c *Config) SetMusicVolume(v float32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.MusicVolume = clamp(v, 0.0, 1.0)
}

func (c *Config) IsMuted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Muted
}

func (c *Config) SetMuted(m bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Muted = m
}

func (c *Config) Is3DAudioEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Enable3DAudio
}

func (c *Config) Set3DAudioEnabled(e bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Enable3DAudio = e
}
