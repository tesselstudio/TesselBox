package audio

import (
	"time"
)

// Sound represents a loaded sound asset
type Sound struct {
	ID     SoundID
	Name   string
	Data   []byte
	Volume float32
}

// SoundID identifies a specific sound effect
type SoundID int

const (
	SFXBlockBreak SoundID = iota
	SFXBlockPlace
	SFXFootstep
	SFXUIClick
	SFXUIHover
	SFXItemPickup
	SFXPlayerDamage
	SFXJump
)

// GameEvent represents game events that trigger sounds
type GameEvent int

const (
	EventBlockBreak GameEvent = iota
	EventBlockPlace
	EventFootstep
	EventUIClick
	EventUIHover
	EventItemPickup
	EventPlayerDamage
	EventJump
)

// Vec3 represents a 3D position for positional audio
type Vec3 struct {
	X, Y, Z float32
}

// MusicTrack represents a music track
type MusicTrack struct {
	Name     string
	Path     string
	Title    string
	Artist   string
	Duration time.Duration
	Mood     MusicMood
}

// MusicMood represents the mood/intensity of music
type MusicMood int

const (
	MoodCalm MusicMood = iota
	MoodExploration
	MoodTension
	MoodCombat
)
