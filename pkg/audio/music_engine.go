package audio

import (
	"fmt"
	"path/filepath"
	"sync"
)

// MusicEngine handles background music playback
type MusicEngine struct {
	mu           sync.RWMutex
	loaded       bool
	basePath     string
	volume       float32
	currentTrack *MusicTrack
	isPlaying    bool
	isPaused     bool
	playlist     []MusicTrack
	currentIndex int
	shuffle      bool
	repeat       bool
}

// NewMusicEngine creates a new music engine
func NewMusicEngine(basePath string) *MusicEngine {
	return &MusicEngine{
		basePath:     basePath,
		volume:       1.0,
		playlist:     make([]MusicTrack, 0),
		currentIndex: -1,
	}
}

// Initialize sets up the music engine
func (e *MusicEngine) Initialize() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.loaded {
		return nil
	}

	// Load default playlist
	e.loadDefaultPlaylist()

	e.loaded = true
	return nil
}

// Shutdown cleans up music resources
func (e *MusicEngine) Shutdown() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.loaded {
		return
	}

	e.Stop()
	e.playlist = nil
	e.loaded = false
}

// Play starts playing a music track
func (e *MusicEngine) Play(trackName string, volume float32) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.loaded {
		return
	}

	// Stop current track if playing
	if e.isPlaying {
		e.stopInternal()
	}

	// Find track
	for i, track := range e.playlist {
		if track.Name == trackName {
			e.currentTrack = &e.playlist[i]
			e.currentIndex = i
			break
		}
	}

	if e.currentTrack == nil {
		// Try to load as a direct file
		path := filepath.Join(e.basePath, "music", trackName)
		e.currentTrack = &MusicTrack{
			Name: trackName,
			Path: path,
		}
	}

	e.volume = volume
	e.isPlaying = true
	e.isPaused = false

	// In a full implementation, start streaming audio playback
	_ = fmt.Sprintf("Playing music: %s at volume %.2f", trackName, volume)
}

// Stop stops the current music
func (e *MusicEngine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.loaded || !e.isPlaying {
		return
	}

	e.stopInternal()
}

// Pause pauses the current music
func (e *MusicEngine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.loaded || !e.isPlaying || e.isPaused {
		return
	}

	e.isPaused = true
	// In a full implementation, pause audio playback
}

// Resume resumes paused music
func (e *MusicEngine) Resume() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.loaded || !e.isPlaying || !e.isPaused {
		return
	}

	e.isPaused = false
	// In a full implementation, resume audio playback
}

// SetVolume sets the music volume
func (e *MusicEngine) SetVolume(volume float32) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.volume = clamp(volume, 0.0, 1.0)

	if e.isPlaying && !e.isPaused {
		// In a full implementation, update playback volume
	}
}

// Update should be called each frame to handle track transitions
func (e *MusicEngine) Update() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.loaded || !e.isPlaying || e.isPaused {
		return
	}

	// In a full implementation:
	// 1. Check if current track finished
	// 2. Auto-advance to next track if needed
	// 3. Handle crossfading
}

// Next plays the next track in the playlist
func (e *MusicEngine) Next() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.playlist) == 0 {
		return
	}

	e.currentIndex++
	if e.currentIndex >= len(e.playlist) {
		if e.repeat {
			e.currentIndex = 0
		} else {
			e.currentIndex = len(e.playlist) - 1
			return
		}
	}

	track := e.playlist[e.currentIndex]
	e.mu.Unlock()
	e.Play(track.Name, e.volume)
}

// Previous plays the previous track in the playlist
func (e *MusicEngine) Previous() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.playlist) == 0 {
		return
	}

	e.currentIndex--
	if e.currentIndex < 0 {
		if e.repeat {
			e.currentIndex = len(e.playlist) - 1
		} else {
			e.currentIndex = 0
			return
		}
	}

	track := e.playlist[e.currentIndex]
	e.mu.Unlock()
	e.Play(track.Name, e.volume)
}

// SetShuffle enables/disables shuffle mode
func (e *MusicEngine) SetShuffle(shuffle bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.shuffle = shuffle
}

// SetRepeat enables/disables repeat mode
func (e *MusicEngine) SetRepeat(repeat bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.repeat = repeat
}

// GetCurrentTrack returns the currently playing track
func (e *MusicEngine) GetCurrentTrack() *MusicTrack {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.currentTrack
}

// IsPlaying returns true if music is playing
func (e *MusicEngine) IsPlaying() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.isPlaying && !e.isPaused
}

// AddToPlaylist adds a track to the playlist
func (e *MusicEngine) AddToPlaylist(track MusicTrack) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.playlist = append(e.playlist, track)
}

// ClearPlaylist clears the playlist
func (e *MusicEngine) ClearPlaylist() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.playlist = make([]MusicTrack, 0)
	e.currentIndex = -1
}

// Private methods

func (e *MusicEngine) stopInternal() {
	e.isPlaying = false
	e.isPaused = false
	e.currentTrack = nil
	// In a full implementation, stop audio playback
}

func (e *MusicEngine) loadDefaultPlaylist() {
	// Define default tracks
	defaults := []MusicTrack{
		{
			Name:   "exploration",
			Path:   filepath.Join(e.basePath, "music", "exploration.ogg"),
			Title:  "Exploration",
			Artist: "TesselBox",
			Mood:   MoodExploration,
		},
		{
			Name:   "calm",
			Path:   filepath.Join(e.basePath, "music", "calm.ogg"),
			Title:  "Calm",
			Artist: "TesselBox",
			Mood:   MoodCalm,
		},
		{
			Name:   "tension",
			Path:   filepath.Join(e.basePath, "music", "tension.ogg"),
			Title:  "Tension",
			Artist: "TesselBox",
			Mood:   MoodTension,
		},
	}

	e.playlist = defaults
}
