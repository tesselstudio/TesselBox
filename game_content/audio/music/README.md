# Background Music

This directory contains background music tracks for TesselBox.

## File Format

Music files should be in OGG Vorbis format (.ogg) for best compression.

## Required Music Files

| File | Description | Mood |
|------|-------------|------|
| `exploration.ogg` | Main exploration theme | Exploration |
| `calm.ogg` | Peaceful/ambient music | Calm |
| `tension.ogg` | Danger/tension music | Tension |

## Music System

The game uses a playlist-based music system with the following features:

- **Shuffle mode**: Random track order
- **Repeat mode**: Loop playlist when finished
- **Crossfade**: Smooth transitions between tracks (not yet implemented)
- **Dynamic mood**: Music changes based on game state (not yet implemented)

## Volume

Default music volume is 0.7 (70% of master volume).

## Integration

Music is controlled by the `MusicEngine` in `pkg/audio/music_engine.go`.
