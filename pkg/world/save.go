package world

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// SaveManager handles saving and loading world data
type SaveManager struct {
	mu sync.RWMutex

	worldName string
	basePath  string

	// Async save queue
	saveQueue chan *Chunk
	wg        sync.WaitGroup
	running   bool
}

// NewSaveManager creates a new save manager
func NewSaveManager(worldName string) (*SaveManager, error) {
	// Sanitize world name
	sanitizedName := filepath.Base(filepath.Clean(worldName))
	if sanitizedName != worldName || sanitizedName == "." || sanitizedName == ".." {
		return nil, fmt.Errorf("invalid world name: %s", worldName)
	}

	basePath := filepath.Join("saves", sanitizedName)

	// Create save directory
	if err := os.MkdirAll(filepath.Join(basePath, "region"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create save directory: %w", err)
	}

	sm := &SaveManager{
		worldName: worldName,
		basePath:  basePath,
		saveQueue: make(chan *Chunk, 100),
		running:   true,
	}

	// Start async save worker
	sm.wg.Add(1)
	go sm.saveWorker()

	return sm, nil
}

// Close shuts down the save manager
func (sm *SaveManager) Close() {
	sm.mu.Lock()
	sm.running = false
	sm.mu.Unlock()

	close(sm.saveQueue)
	sm.wg.Wait()
}

// SaveChunk saves a chunk to disk (async)
func (sm *SaveManager) SaveChunk(chunk *Chunk) {
	sm.mu.RLock()
	if !sm.running {
		sm.mu.RUnlock()
		return
	}
	sm.mu.RUnlock()

	select {
	case sm.saveQueue <- chunk:
		// Queued for save
	default:
		// Queue full, save synchronously
		sm.saveChunkSync(chunk)
	}
}

// saveWorker processes the save queue
func (sm *SaveManager) saveWorker() {
	defer sm.wg.Done()

	for chunk := range sm.saveQueue {
		sm.saveChunkSync(chunk)
	}
}

// saveChunkSync saves a chunk synchronously
func (sm *SaveManager) saveChunkSync(chunk *Chunk) error {
	// Serialize chunk data
	data, err := sm.serializeChunk(chunk)
	if err != nil {
		return err
	}

	// Get region file path
	regionPath := sm.getRegionPath(chunk.Coord)

	// Ensure region directory exists
	regionDir := filepath.Dir(regionPath)
	if err := os.MkdirAll(regionDir, 0755); err != nil {
		return err
	}

	// Write to temp file first
	tempPath := regionPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}

	// Atomic rename
	if err := os.Rename(tempPath, regionPath); err != nil {
		return err
	}

	chunk.MarkSaved()
	return nil
}

// LoadChunk loads a chunk from disk
func (sm *SaveManager) LoadChunk(coord ChunkCoord) (*Chunk, error) {
	regionPath := sm.getRegionPath(coord)

	// Check if file exists
	if _, err := os.Stat(regionPath); os.IsNotExist(err) {
		return nil, nil // Chunk doesn't exist yet
	}

	// Read data
	data, err := os.ReadFile(regionPath)
	if err != nil {
		return nil, err
	}

	// Deserialize
	return sm.deserializeChunk(coord, data)
}

// serializeChunk serializes chunk data to bytes
func (sm *SaveManager) serializeChunk(chunk *Chunk) ([]byte, error) {
	chunk.mu.RLock()
	defer chunk.mu.RUnlock()

	var buf bytes.Buffer

	// Write header
	// Version: 1 byte
	buf.WriteByte(1)

	// Chunk coordinates: 8 bytes
	binary.Write(&buf, binary.LittleEndian, int32(chunk.Coord.X))
	binary.Write(&buf, binary.LittleEndian, int32(chunk.Coord.Z))

	// Biome: 1 byte
	binary.Write(&buf, binary.LittleEndian, uint8(chunk.biome))

	// Heightmap: 256 * 2 bytes = 512 bytes
	for x := 0; x < ChunkSize; x++ {
		for z := 0; z < ChunkSize; z++ {
			binary.Write(&buf, binary.LittleEndian, chunk.heightmap[x][z])
		}
	}

	// Block data with RLE compression
	// Count non-air blocks first
	blockCount := 0
	for _, block := range chunk.blocks {
		if block.ID != BlockIDAir {
			blockCount++
		}
	}

	binary.Write(&buf, binary.LittleEndian, int32(blockCount))

	// Write block data
	for i, block := range chunk.blocks {
		if block.ID != BlockIDAir {
			// Position: 2 bytes each (x, y, z)
			x := i % ChunkSize
			z := (i / ChunkSize) % ChunkSize
			y := i / (ChunkSize * ChunkSize)

			binary.Write(&buf, binary.LittleEndian, uint16(x))
			binary.Write(&buf, binary.LittleEndian, uint16(y))
			binary.Write(&buf, binary.LittleEndian, uint16(z))

			// Block data: 4 bytes
			binary.Write(&buf, binary.LittleEndian, block.ID)
			binary.Write(&buf, binary.LittleEndian, block.Metadata)
			binary.Write(&buf, binary.LittleEndian, block.Light)
		}
	}

	// Compress with gzip
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	gzipWriter.Write(buf.Bytes())
	gzipWriter.Close()

	return compressed.Bytes(), nil
}

// deserializeChunk deserializes chunk data from bytes
func (sm *SaveManager) deserializeChunk(coord ChunkCoord, data []byte) (*Chunk, error) {
	// Decompress
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(gzipReader); err != nil {
		return nil, err
	}

	reader := bytes.NewReader(buf.Bytes())

	// Read header
	version, _ := reader.ReadByte()
	if version != 1 {
		return nil, fmt.Errorf("unsupported chunk version: %d", version)
	}

	var chunkCoord ChunkCoord
	var biome uint8

	// Read coordinates
	var x, z int32
	binary.Read(reader, binary.LittleEndian, &x)
	binary.Read(reader, binary.LittleEndian, &z)
	chunkCoord.X = int(x)
	chunkCoord.Z = int(z)

	// Read biome
	binary.Read(reader, binary.LittleEndian, &biome)

	// Create chunk
	chunk := NewChunk(chunkCoord)
	chunk.SetBiome(BiomeType(biome))

	// Read heightmap
	for i := 0; i < ChunkSize; i++ {
		for j := 0; j < ChunkSize; j++ {
			binary.Read(reader, binary.LittleEndian, &chunk.heightmap[i][j])
		}
	}

	// Read block data
	var blockCount int32
	binary.Read(reader, binary.LittleEndian, &blockCount)

	for i := int32(0); i < blockCount; i++ {
		var bx, by, bz uint16
		var id BlockID
		var metadata, light uint8

		binary.Read(reader, binary.LittleEndian, &bx)
		binary.Read(reader, binary.LittleEndian, &by)
		binary.Read(reader, binary.LittleEndian, &bz)
		binary.Read(reader, binary.LittleEndian, &id)
		binary.Read(reader, binary.LittleEndian, &metadata)
		binary.Read(reader, binary.LittleEndian, &light)

		chunk.SetBlock(int(bx), int(by), int(bz), BlockData{
			ID:       id,
			Metadata: metadata,
			Light:    light,
		})
	}

	chunk.modified = false
	chunk.meshDirty = true

	return chunk, nil
}

// getRegionPath returns the file path for a chunk
func (sm *SaveManager) getRegionPath(coord ChunkCoord) string {
	// For simplicity, store each chunk as a separate file
	chunkFile := fmt.Sprintf("c.%d.%d.dat", coord.X, coord.Z)
	return filepath.Join(sm.basePath, "region", chunkFile)
}

// SaveWorldInfo saves world metadata
func (sm *SaveManager) SaveWorldInfo(info WorldInfo) error {
	infoPath := filepath.Join(sm.basePath, "world.dat")

	var buf bytes.Buffer

	// Version
	buf.WriteByte(1)

	// World name length and string
	binary.Write(&buf, binary.LittleEndian, int32(len(info.Name)))
	buf.WriteString(info.Name)

	// Seed
	binary.Write(&buf, binary.LittleEndian, info.Seed)

	// Last played
	binary.Write(&buf, binary.LittleEndian, info.LastPlayed)

	// Game time
	binary.Write(&buf, binary.LittleEndian, int32(info.GameTime))

	// Spawn point
	binary.Write(&buf, binary.LittleEndian, info.SpawnX)
	binary.Write(&buf, binary.LittleEndian, info.SpawnY)
	binary.Write(&buf, binary.LittleEndian, info.SpawnZ)

	// Compress
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	gzipWriter.Write(buf.Bytes())
	gzipWriter.Close()

	return os.WriteFile(infoPath, compressed.Bytes(), 0644)
}

// LoadWorldInfo loads world metadata
func (sm *SaveManager) LoadWorldInfo() (*WorldInfo, error) {
	infoPath := filepath.Join(sm.basePath, "world.dat")

	data, err := os.ReadFile(infoPath)
	if err != nil {
		return nil, err
	}

	// Decompress
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(gzipReader); err != nil {
		return nil, err
	}

	reader := bytes.NewReader(buf.Bytes())

	// Version
	version, _ := reader.ReadByte()
	if version != 1 {
		return nil, fmt.Errorf("unsupported world version: %d", version)
	}

	var info WorldInfo

	// World name
	var nameLen int32
	binary.Read(reader, binary.LittleEndian, &nameLen)
	nameBytes := make([]byte, nameLen)
	reader.Read(nameBytes)
	info.Name = string(nameBytes)

	// Other fields
	binary.Read(reader, binary.LittleEndian, &info.Seed)
	binary.Read(reader, binary.LittleEndian, &info.LastPlayed)

	var gameTime int32
	binary.Read(reader, binary.LittleEndian, &gameTime)
	info.GameTime = int(gameTime)

	binary.Read(reader, binary.LittleEndian, &info.SpawnX)
	binary.Read(reader, binary.LittleEndian, &info.SpawnY)
	binary.Read(reader, binary.LittleEndian, &info.SpawnZ)

	return &info, nil
}

// ListWorlds returns a list of available worlds
func ListWorlds() ([]string, error) {
	entries, err := os.ReadDir("saves")
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	worlds := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			worlds = append(worlds, entry.Name())
		}
	}

	return worlds, nil
}

// WorldExists checks if a world exists
func WorldExists(name string) bool {
	// Sanitize world name
	sanitizedName := filepath.Base(filepath.Clean(name))
	if sanitizedName != name || sanitizedName == "." || sanitizedName == ".." {
		return false
	}

	path := filepath.Join("saves", sanitizedName)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// DeleteWorld deletes a world
func DeleteWorld(name string) error {
	// Sanitize world name
	sanitizedName := filepath.Base(filepath.Clean(name))
	if sanitizedName != name || sanitizedName == "." || sanitizedName == ".." {
		return fmt.Errorf("invalid world name: %s", name)
	}

	path := filepath.Join("saves", sanitizedName)
	return os.RemoveAll(path)
}
