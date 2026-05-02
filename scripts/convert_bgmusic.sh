#!/bin/bash
# Convert bgmusic.ly to WAV format for use in the game

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LILYPOND_FILE="$PROJECT_ROOT/bgmusic.ly"
OUTPUT_DIR="$PROJECT_ROOT/pkg/audio/assets/music"
OUTPUT_FILE="$OUTPUT_DIR/gameplay_music.wav"

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

echo "Converting bgmusic.ly to WAV format..."
echo "Input: $LILYPOND_FILE"
echo "Output: $OUTPUT_FILE"

# Check if lilypond is installed
if ! command -v lilypond &> /dev/null; then
    echo "Error: lilypond is not installed"
    echo "Install it with: sudo apt-get install lilypond (Ubuntu/Debian)"
    echo "              or: brew install lilypond (macOS)"
    exit 1
fi

# Check if timidity is installed (for MIDI to WAV conversion)
if ! command -v timidity &> /dev/null; then
    echo "Error: timidity is not installed"
    echo "Install it with: sudo apt-get install timidity (Ubuntu/Debian)"
    echo "              or: brew install timidity (macOS)"
    exit 1
fi

# Convert LilyPond to MIDI
echo "Step 1: Converting LilyPond to MIDI..."
MIDI_FILE="/tmp/bgmusic.midi"
lilypond --output=/tmp --format=midi "$LILYPOND_FILE" 2>/dev/null || true

# Find the generated MIDI file
if [ ! -f "$MIDI_FILE" ]; then
    # Try alternative naming
    MIDI_FILE="/tmp/bgmusic.mid"
    if [ ! -f "$MIDI_FILE" ]; then
        echo "Error: Failed to generate MIDI file from LilyPond"
        exit 1
    fi
fi

echo "Step 2: Converting MIDI to WAV..."
# Convert MIDI to WAV with timidity
timidity "$MIDI_FILE" -Ow -o "$OUTPUT_FILE" 2>/dev/null || {
    echo "Error: Failed to convert MIDI to WAV"
    exit 1
}

# Verify the output file was created
if [ ! -f "$OUTPUT_FILE" ]; then
    echo "Error: Output WAV file was not created"
    exit 1
fi

# Get file size
FILE_SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
echo "✓ Successfully created: $OUTPUT_FILE ($FILE_SIZE)"

# Clean up temporary files
rm -f "$MIDI_FILE" /tmp/bgmusic.* 2>/dev/null || true

echo "Done! The background music is ready to use."
echo ""
echo "The game will automatically load this as 'gameplay_music' and loop it."
