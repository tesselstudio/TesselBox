// Package blocks provides block type definitions, registry management,
// and geometry generation for hexagonal prism blocks.
//
// The block system is responsible for:
//   - Defining block types and their properties (solid, transparent, etc.)
//   - Managing the global block registry
//   - Generating mesh geometry for rendering
//   - Handling block placement and removal
//
// # Block Types
//
// The following block types are defined:
//   - BlockTypeFull: Full hexagonal prism block
//   - BlockTypeHalfVertical: Vertical half-height block
//   - BlockTypeHalfHorizontal: Horizontal half-height block
//   - BlockTypeCorner: Corner wedge block
//   - BlockTypeStairs: Stair block
//   - BlockTypeSlab: Slab block
//
// Usage
//
//	registry := blocks.NewBlockRegistry()
//	block, exists := registry.GetBlock("stone")
//	if exists {
//	    fmt.Println(block.Name) // "Stone"
//	}
//
// # Thread Safety
//
// The block registry is thread-safe for concurrent reads after initialization.
// All modifications should be done during initialization before the game starts.
package blocks

import (
	"strconv"
	"sync"
)

// Color represents a simple RGBA color
type Color struct {
	R, G, B, A uint8
}

// Float represents a simple float type
type Float float32

// ColorGray returns a gray color
func ColorGray() Color {
	return Color{128, 128, 128, 255}
}

// NewColor creates a new color from RGBA values
func NewColor(r, g, b, a uint8) Color {
	return Color{r, g, b, a}
}

// BlockType represents different types of hexagonal blocks
type BlockType int

const (
	BlockTypeFull BlockType = iota
	BlockTypeHalfVertical
	BlockTypeHalfHorizontal
	BlockTypeCorner
	BlockTypeStairs
	BlockTypeSlab
)

// BlockProperties defines the properties of a block type
type BlockProperties struct {
	ID          string
	Name        string
	Type        BlockType
	Solid       bool
	Transparent bool
	LightLevel  int
	TintColor   Color
	Mass        Float
	Resistance  Float
}

// BlockRegistry manages all registered block types
type BlockRegistry struct {
	mu     sync.RWMutex
	blocks map[string]*BlockProperties
	byID   map[BlockType]*BlockProperties
	nextID int
}

// NewBlockRegistry creates a new block registry
func NewBlockRegistry() *BlockRegistry {
	registry := &BlockRegistry{
		blocks: make(map[string]*BlockProperties),
		byID:   make(map[BlockType]*BlockProperties),
		nextID: 0,
	}

	// Register default block types
	registry.registerDefaults()

	return registry
}

// RegisterBlock registers a new block type
func (r *BlockRegistry) RegisterBlock(props *BlockProperties) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.blocks[props.ID]; exists {
		return ErrBlockAlreadyExists
	}

	// Assign unique ID if not set
	if props.ID == "" {
		props.ID = r.generateID()
	}

	r.blocks[props.ID] = props
	r.byID[props.Type] = props

	return nil
}

// GetBlock retrieves a block by ID
func (r *BlockRegistry) GetBlock(id string) (*BlockProperties, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	block, exists := r.blocks[id]
	return block, exists
}

// GetBlockByType retrieves a block by type
func (r *BlockRegistry) GetBlockByType(blockType BlockType) (*BlockProperties, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	block, exists := r.byID[blockType]
	return block, exists
}

// GetAllBlocks returns all registered blocks
func (r *BlockRegistry) GetAllBlocks() []*BlockProperties {
	r.mu.RLock()
	defer r.mu.RUnlock()

	blocks := make([]*BlockProperties, 0, len(r.blocks))
	for _, block := range r.blocks {
		blocks = append(blocks, block)
	}

	return blocks
}

// GetSolidBlocks returns all solid blocks
func (r *BlockRegistry) GetSolidBlocks() []*BlockProperties {
	r.mu.RLock()
	defer r.mu.RUnlock()

	blocks := make([]*BlockProperties, 0)
	for _, block := range r.blocks {
		if block.Solid {
			blocks = append(blocks, block)
		}
	}

	return blocks
}

// GetTransparentBlocks returns all transparent blocks
func (r *BlockRegistry) GetTransparentBlocks() []*BlockProperties {
	r.mu.RLock()
	defer r.mu.RUnlock()

	blocks := make([]*BlockProperties, 0)
	for _, block := range r.blocks {
		if block.Transparent {
			blocks = append(blocks, block)
		}
	}

	return blocks
}

// registerDefaults registers the default block types
func (r *BlockRegistry) registerDefaults() {
	defaultBlocks := []*BlockProperties{
		{
			ID:          "stone",
			Name:        "Stone",
			Type:        BlockTypeFull,
			Solid:       true,
			Transparent: false,
			LightLevel:  0,
			TintColor:   ColorGray(),
			Mass:        3.0,
			Resistance:  6.0,
		},
		{
			ID:          "dirt",
			Name:        "Dirt",
			Type:        BlockTypeFull,
			Solid:       true,
			Transparent: false,
			LightLevel:  0,
			TintColor:   NewColor(139, 90, 43, 255),
			Mass:        2.5,
			Resistance:  2.5,
		},
		{
			ID:          "grass",
			Name:        "Grass",
			Type:        BlockTypeFull,
			Solid:       true,
			Transparent: false,
			LightLevel:  0,
			TintColor:   NewColor(124, 252, 0, 255),
			Mass:        2.0,
			Resistance:  3.0,
		},
		{
			ID:          "wood",
			Name:        "Wood",
			Type:        BlockTypeFull,
			Solid:       true,
			Transparent: false,
			LightLevel:  0,
			TintColor:   NewColor(139, 69, 19, 255),
			Mass:        1.5,
			Resistance:  4.0,
		},
		{
			ID:          "glass",
			Name:        "Glass",
			Type:        BlockTypeFull,
			Solid:       true,
			Transparent: true,
			LightLevel:  0,
			TintColor:   NewColor(200, 200, 255, 128),
			Mass:        2.5,
			Resistance:  1.5,
		},
		{
			ID:          "water",
			Name:        "Water",
			Type:        BlockTypeFull,
			Solid:       false,
			Transparent: true,
			LightLevel:  15,
			TintColor:   NewColor(64, 164, 223, 180),
			Mass:        1.0,
			Resistance:  0.5,
		},
		{
			ID:          "half_vertical",
			Name:        "Half Block (Vertical)",
			Type:        BlockTypeHalfVertical,
			Solid:       true,
			Transparent: false,
			LightLevel:  0,
			TintColor:   ColorGray(),
			Mass:        1.5,
			Resistance:  3.0,
		},
		{
			ID:          "half_horizontal",
			Name:        "Half Block (Horizontal)",
			Type:        BlockTypeHalfHorizontal,
			Solid:       true,
			Transparent: false,
			LightLevel:  0,
			TintColor:   ColorGray(),
			Mass:        1.5,
			Resistance:  3.0,
		},
		{
			ID:          "corner",
			Name:        "Corner Block",
			Type:        BlockTypeCorner,
			Solid:       true,
			Transparent: false,
			LightLevel:  0,
			TintColor:   ColorGray(),
			Mass:        1.0,
			Resistance:  3.0,
		},
		{
			ID:          "stairs",
			Name:        "Stairs",
			Type:        BlockTypeStairs,
			Solid:       true,
			Transparent: false,
			LightLevel:  0,
			TintColor:   ColorGray(),
			Mass:        2.0,
			Resistance:  4.0,
		},
		{
			ID:          "slab",
			Name:        "Slab",
			Type:        BlockTypeSlab,
			Solid:       true,
			Transparent: false,
			LightLevel:  0,
			TintColor:   ColorGray(),
			Mass:        1.0,
			Resistance:  2.0,
		},
	}

	for _, block := range defaultBlocks {
		r.blocks[block.ID] = block
		r.byID[block.Type] = block
	}
}

// generateID generates a unique block ID
func (r *BlockRegistry) generateID() string {
	id := r.nextID
	r.nextID++
	return "block_" + strconv.Itoa(id)
}

// Global block registry instance
var GlobalRegistry = NewBlockRegistry()

// GetGlobalRegistry returns the global block registry
func GetGlobalRegistry() *BlockRegistry {
	return GlobalRegistry
}

// Errors
var (
	ErrBlockAlreadyExists = &BlockError{Message: "block already exists"}
	ErrBlockNotFound      = &BlockError{Message: "block not found"}
	ErrInvalidBlockType   = &BlockError{Message: "invalid block type"}
)

// BlockError represents a block-related error
type BlockError struct {
	Message string
}

func (e *BlockError) Error() string {
	return e.Message
}
