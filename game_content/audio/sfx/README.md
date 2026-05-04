# Sound Effects

This directory contains sound effect files for TesselBox.

## File Format

Sound effects should be in OGG Vorbis format (.ogg) for best compression and quality.
Alternative: WAV format (.wav) for uncompressed audio.

## Required Sound Files

| File | Description | Event |
|------|-------------|-------|
| `block_break.ogg` | Block breaking sound | BlockBreak |
| `block_place.ogg` | Block placement sound | BlockPlace |
| `footstep.ogg` | Player footstep | Footstep |
| `ui_click.ogg` | UI button click | UIClick |
| `ui_hover.ogg` | UI hover sound | UIHover |
| `item_pickup.ogg` | Item pickup sound | ItemPickup |
| `player_damage.ogg` | Player taking damage | PlayerDamage |
| `jump.ogg` | Player jump sound | Jump |

## Volume Guidelines

- UI sounds: 0.4 - 0.8 (not too loud)
- Gameplay sounds: 0.6 - 1.0
- Ambient sounds: 0.3 - 0.6

## Generating Placeholder Sounds

You can generate simple placeholder sounds using tools like:
- sfxr (http://www.drpetter.se/project_sfxr.html)
- bfxr (https://www.bfxr.net/)
- Audacity (https://www.audacityteam.org/)

## Integration

Sounds are automatically loaded by the audio manager at initialization.
Missing sounds are logged as warnings but won't crash the game.
