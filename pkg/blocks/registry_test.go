package blocks

import (
	"testing"
)

func TestNewBlockRegistry(t *testing.T) {
	registry := NewBlockRegistry()
	if registry == nil {
		t.Fatal("NewBlockRegistry() returned nil")
	}
}

func TestGetBlock(t *testing.T) {
	registry := NewBlockRegistry()

	// Test getting valid block by ID
	block, exists := registry.GetBlock("stone")
	if !exists {
		t.Error("GetBlock('stone') returned false for existing block")
	}
	if block == nil {
		t.Error("GetBlock('stone') returned nil")
	}
	if block.Name != "Stone" {
		t.Errorf("Expected block name 'Stone', got '%s'", block.Name)
	}

	// Test getting invalid block by ID
	block, exists = registry.GetBlock("nonexistent")
	if exists {
		t.Error("GetBlock('nonexistent') returned true for non-existent block")
	}
	if block != nil {
		t.Error("GetBlock('nonexistent') should return nil")
	}
}

func TestGetBlockByType(t *testing.T) {
	registry := NewBlockRegistry()

	// Test getting valid block by type
	block, exists := registry.GetBlockByType(BlockTypeFull)
	if !exists {
		t.Error("GetBlockByType(BlockTypeFull) returned false for existing block")
	}
	if block == nil {
		t.Error("GetBlockByType(BlockTypeFull) returned nil")
	}

	// Test getting invalid block by type
	block, exists = registry.GetBlockByType(BlockType(999))
	if exists {
		t.Error("GetBlockByType(invalid) returned true for non-existent block")
	}
	if block != nil {
		t.Error("GetBlockByType(invalid) should return nil")
	}
}

func TestGetAllBlocks(t *testing.T) {
	registry := NewBlockRegistry()
	blocks := registry.GetAllBlocks()

	if len(blocks) == 0 {
		t.Error("GetAllBlocks() returned empty slice")
	}

	// Verify no duplicates
	seen := make(map[BlockType]bool)
	for _, block := range blocks {
		if seen[block.Type] {
			t.Errorf("Duplicate block type: %d", block.Type)
		}
		seen[block.Type] = true
	}
}

func TestBlockProperties(t *testing.T) {
	registry := NewBlockRegistry()

	// Test stone properties
	stone, exists := registry.GetBlock("stone")
	if !exists {
		t.Fatal("Stone block not found in registry")
	}
	if !stone.Solid {
		t.Error("Stone should be solid")
	}
	if stone.Transparent {
		t.Error("Stone should not be transparent")
	}

	// Test air properties (if it exists)
	air, exists := registry.GetBlock("air")
	if exists {
		if air.Solid {
			t.Error("Air should not be solid")
		}
		if !air.Transparent {
			t.Error("Air should be transparent")
		}
	}
}

func TestGenerateID(t *testing.T) {
	registry := NewBlockRegistry()

	// Test ID generation
	id1 := registry.generateID()
	id2 := registry.generateID()

	if id1 == id2 {
		t.Error("generateID() should produce unique IDs")
	}
	if id1 != "block_0" {
		t.Errorf("Expected first ID to be 'block_0', got '%s'", id1)
	}
	if id2 != "block_1" {
		t.Errorf("Expected second ID to be 'block_1', got '%s'", id2)
	}
}

func TestRegisterBlock(t *testing.T) {
	registry := NewBlockRegistry()

	// Test registering new block
	newBlock := &BlockProperties{
		ID:    "test_block",
		Name:  "Test Block",
		Type:  BlockType(999),
		Solid: true,
	}

	err := registry.RegisterBlock(newBlock)
	if err != nil {
		t.Errorf("RegisterBlock() returned error: %v", err)
	}

	// Verify block was registered
	block, exists := registry.GetBlock("test_block")
	if !exists {
		t.Error("Registered block not found")
	}
	if block.Name != "Test Block" {
		t.Errorf("Expected block name 'Test Block', got '%s'", block.Name)
	}

	// Test duplicate registration
	err = registry.RegisterBlock(newBlock)
	if err == nil {
		t.Error("RegisterBlock() should return error for duplicate block")
	}
}
