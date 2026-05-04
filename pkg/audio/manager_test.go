package audio

import (
	"testing"
)

func TestGetManager(t *testing.T) {
	m1 := GetManager()
	m2 := GetManager()

	if m1 == nil {
		t.Error("GetManager() returned nil")
	}

	if m1 != m2 {
		t.Error("GetManager() should return the same singleton instance")
	}
}

func TestManagerVolumes(t *testing.T) {
	m := GetManager()

	// Test default volumes
	if m.GetMasterVolume() != 1.0 {
		t.Errorf("Expected master volume 1.0, got %f", m.GetMasterVolume())
	}

	if m.GetSFXVolume() != 1.0 {
		t.Errorf("Expected SFX volume 1.0, got %f", m.GetSFXVolume())
	}

	if m.GetMusicVolume() != 0.7 {
		t.Errorf("Expected music volume 0.7, got %f", m.GetMusicVolume())
	}

	// Test setting volumes
	m.SetMasterVolume(0.5)
	if m.GetMasterVolume() != 0.5 {
		t.Errorf("Expected master volume 0.5, got %f", m.GetMasterVolume())
	}

	m.SetSFXVolume(0.8)
	if m.GetSFXVolume() != 0.8 {
		t.Errorf("Expected SFX volume 0.8, got %f", m.GetSFXVolume())
	}

	m.SetMusicVolume(0.3)
	if m.GetMusicVolume() != 0.3 {
		t.Errorf("Expected music volume 0.3, got %f", m.GetMusicVolume())
	}

	// Test volume clamping
	m.SetMasterVolume(2.0)
	if m.GetMasterVolume() != 1.0 {
		t.Errorf("Expected clamped master volume 1.0, got %f", m.GetMasterVolume())
	}

	m.SetMasterVolume(-0.5)
	if m.GetMasterVolume() != 0.0 {
		t.Errorf("Expected clamped master volume 0.0, got %f", m.GetMasterVolume())
	}
}

func TestManagerMute(t *testing.T) {
	m := GetManager()

	// Test default mute state
	if m.IsMuted() {
		t.Error("Expected muted to be false by default")
	}

	// Test setting muted
	m.SetMuted(true)
	if !m.IsMuted() {
		t.Error("Expected muted to be true after SetMuted(true)")
	}

	m.SetMuted(false)
	if m.IsMuted() {
		t.Error("Expected muted to be false after SetMuted(false)")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MasterVolume != 1.0 {
		t.Errorf("Expected default master volume 1.0, got %f", config.MasterVolume)
	}

	if config.SFXVolume != 1.0 {
		t.Errorf("Expected default SFX volume 1.0, got %f", config.SFXVolume)
	}

	if config.MusicVolume != 0.7 {
		t.Errorf("Expected default music volume 0.7, got %f", config.MusicVolume)
	}

	if config.Muted {
		t.Error("Expected muted to be false by default")
	}

	if !config.Enable3DAudio {
		t.Error("Expected 3D audio to be enabled by default")
	}

	if !config.HighQuality {
		t.Error("Expected high quality to be enabled by default")
	}
}

func TestConfigVolumeSetters(t *testing.T) {
	config := DefaultConfig()

	config.SetMasterVolume(0.5)
	if config.GetMasterVolume() != 0.5 {
		t.Errorf("Expected master volume 0.5, got %f", config.GetMasterVolume())
	}

	config.SetSFXVolume(0.6)
	if config.GetSFXVolume() != 0.6 {
		t.Errorf("Expected SFX volume 0.6, got %f", config.GetSFXVolume())
	}

	config.SetMusicVolume(0.4)
	if config.GetMusicVolume() != 0.4 {
		t.Errorf("Expected music volume 0.4, got %f", config.GetMusicVolume())
	}

	// Test clamping
	config.SetMasterVolume(1.5)
	if config.GetMasterVolume() != 1.0 {
		t.Errorf("Expected clamped master volume 1.0, got %f", config.GetMasterVolume())
	}

	config.SetMasterVolume(-0.5)
	if config.GetMasterVolume() != 0.0 {
		t.Errorf("Expected clamped master volume 0.0, got %f", config.GetMasterVolume())
	}
}

func TestConfigMute(t *testing.T) {
	config := DefaultConfig()

	if config.IsMuted() {
		t.Error("Expected muted to be false by default")
	}

	config.SetMuted(true)
	if !config.IsMuted() {
		t.Error("Expected muted to be true after SetMuted(true)")
	}

	config.SetMuted(false)
	if config.IsMuted() {
		t.Error("Expected muted to be false after SetMuted(false)")
	}
}

func TestSoundRegistry(t *testing.T) {
	registry := NewSoundRegistry()

	// Test that non-existent sounds return nil
	if registry.Get(SFXBlockBreak) != nil {
		t.Error("Expected nil for non-existent sound")
	}

	// Test registering a sound
	sound := &Sound{
		Name:   "test",
		Volume: 1.0,
	}
	registry.Register(SFXBlockBreak, sound)

	retrieved := registry.Get(SFXBlockBreak)
	if retrieved == nil {
		t.Error("Expected to retrieve registered sound")
	}

	if retrieved.Name != "test" {
		t.Errorf("Expected sound name 'test', got '%s'", retrieved.Name)
	}

	// Test unloading
	registry.Unload(SFXBlockBreak)
	if registry.Get(SFXBlockBreak) != nil {
		t.Error("Expected nil after unloading sound")
	}
}

func TestPositionalAudio(t *testing.T) {
	listener := Vec3{X: 0, Y: 0, Z: 0}

	// Test attenuation at same position (should be full volume)
	attenuation := CalculateAttenuation(listener, listener)
	if attenuation != 1.0 {
		t.Errorf("Expected attenuation 1.0 at same position, got %f", attenuation)
	}

	// Test attenuation at distance
	soundPos := Vec3{X: 10, Y: 0, Z: 0}
	attenuation = CalculateAttenuation(listener, soundPos)
	if attenuation >= 1.0 || attenuation <= 0 {
		t.Errorf("Expected attenuation between 0 and 1 at distance, got %f", attenuation)
	}

	// Test attenuation at max distance (should be 0)
	farPos := Vec3{X: 100, Y: 0, Z: 0}
	attenuation = CalculateAttenuation(listener, farPos)
	if attenuation != 0 {
		t.Errorf("Expected attenuation 0 at max distance, got %f", attenuation)
	}

	// Test pan calculation
	leftPos := Vec3{X: -5, Y: 0, Z: 0}
	pan := CalculatePan(listener, leftPos)
	if pan >= 0 {
		t.Errorf("Expected negative pan for left position, got %f", pan)
	}

	rightPos := Vec3{X: 5, Y: 0, Z: 0}
	pan = CalculatePan(listener, rightPos)
	if pan <= 0 {
		t.Errorf("Expected positive pan for right position, got %f", pan)
	}

	centerPos := Vec3{X: 0, Y: 0, Z: 5}
	pan = CalculatePan(listener, centerPos)
	if pan != 0 {
		t.Errorf("Expected pan 0 for center position, got %f", pan)
	}
}

func TestDistance(t *testing.T) {
	a := Vec3{X: 0, Y: 0, Z: 0}
	b := Vec3{X: 3, Y: 4, Z: 0}

	dist := Distance(a, b)
	if dist != 5.0 {
		t.Errorf("Expected distance 5.0 (3-4-5 triangle), got %f", dist)
	}
}
