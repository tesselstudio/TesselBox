package crafting

import (
	"sync"
	"kaijuengine.com/matrix"
)

// ItemType represents different types of items
type ItemType int

const (
	ItemTypeBlock ItemType = iota
	ItemTypeTool
	ItemTypeWeapon
	ItemTypeArmor
	ItemTypeFood
	ItemTypeMaterial
	ItemTypeConsumable
)

// Item represents an item in the game
type Item struct {
	ID          string
	Name        string
	Type        ItemType
	Description string
	MaxStack    int
	Durability  int
	Weight      matrix.Float
	Value       int
	Tags        []string
	Properties  map[string]interface{}
}

// ItemStack represents a stack of items
type ItemStack struct {
	Item     *Item
	Quantity int
	Damage   int // Current damage/durability
}

// NewItemStack creates a new item stack
func NewItemStack(item *Item, quantity int) *ItemStack {
	return &ItemStack{
		Item:     item,
		Quantity: quantity,
		Damage:   0,
	}
}

// CanStack checks if this stack can be combined with another
func (is *ItemStack) CanStack(other *ItemStack) bool {
	if is.Item.ID != other.Item.ID {
		return false
	}
	if is.Damage != other.Damage {
		return false
	}
	return is.Quantity+other.Quantity <= is.Item.MaxStack
}

// Stack combines this stack with another
func (is *ItemStack) Stack(other *ItemStack) bool {
	if !is.CanStack(other) {
		return false
	}
	
	is.Quantity += other.Quantity
	return true
}

// Split splits this stack into two
func (is *ItemStack) Split(quantity int) *ItemStack {
	if quantity <= 0 || quantity >= is.Quantity {
		return nil
	}
	
	split := NewItemStack(is.Item, quantity)
	split.Damage = is.Damage
	is.Quantity -= quantity
	
	return split
}

// IsEmpty returns true if the stack is empty
func (is *ItemStack) IsEmpty() bool {
	return is.Quantity <= 0
}

// IsFull returns true if the stack is at max capacity
func (is *ItemStack) IsFull() bool {
	return is.Quantity >= is.Item.MaxStack
}

// ItemRegistry manages all registered items
type ItemRegistry struct {
	mu     sync.RWMutex
	items  map[string]*Item
	byType map[ItemType][]*Item
}

// NewItemRegistry creates a new item registry
func NewItemRegistry() *ItemRegistry {
	registry := &ItemRegistry{
		items:  make(map[string]*Item),
		byType: make(map[ItemType][]*Item),
	}
	
	// Register default items
	registry.registerDefaults()
	
	return registry
}

// RegisterItem registers a new item
func (ir *ItemRegistry) RegisterItem(item *Item) error {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	
	if _, exists := ir.items[item.ID]; exists {
		return ErrItemAlreadyExists
	}
	
	ir.items[item.ID] = item
	ir.byType[item.Type] = append(ir.byType[item.Type], item)
	
	return nil
}

// GetItem retrieves an item by ID
func (ir *ItemRegistry) GetItem(id string) (*Item, bool) {
	ir.mu.RLock()
	defer ir.mu.RUnlock()
	
	item, exists := ir.items[id]
	return item, exists
}

// GetItemsByType retrieves all items of a specific type
func (ir *ItemRegistry) GetItemsByType(itemType ItemType) []*Item {
	ir.mu.RLock()
	defer ir.mu.RUnlock()
	
	items := make([]*Item, len(ir.byType[itemType]))
	copy(items, ir.byType[itemType])
	
	return items
}

// GetAllItems returns all registered items
func (ir *ItemRegistry) GetAllItems() []*Item {
	ir.mu.RLock()
	defer ir.mu.RUnlock()
	
	items := make([]*Item, 0, len(ir.items))
	for _, item := range ir.items {
		items = append(items, item)
	}
	
	return items
}

// registerDefaults registers the default items
func (ir *ItemRegistry) registerDefaults() {
	defaultItems := []*Item{
		// Blocks
		{
			ID:          "stone",
			Name:        "Stone",
			Type:        ItemTypeBlock,
			Description: "A solid stone block",
			MaxStack:    64,
			Durability:  0,
			Weight:      3.0,
			Value:       1,
			Tags:        []string{"block", "stone", "building"},
		},
		{
			ID:          "wood",
			Name:        "Wood",
			Type:        ItemTypeBlock,
			Description: "A wooden block",
			MaxStack:    64,
			Durability:  0,
			Weight:      1.5,
			Value:       2,
			Tags:        []string{"block", "wood", "building"},
		},
		{
			ID:          "iron_ore",
			Name:        "Iron Ore",
			Type:        ItemTypeMaterial,
			Description: "Raw iron ore",
			MaxStack:    64,
			Durability:  0,
			Weight:      2.0,
			Value:       5,
			Tags:        []string{"ore", "metal", "material"},
		},
		{
			ID:          "iron_ingot",
			Name:        "Iron Ingot",
			Type:        ItemTypeMaterial,
			Description: "Refined iron ingot",
			MaxStack:    64,
			Durability:  0,
			Weight:      1.0,
			Value:       10,
			Tags:        []string{"ingot", "metal", "material"},
		},
		{
			ID:          "diamond",
			Name:        "Diamond",
			Type:        ItemTypeMaterial,
			Description: "A precious diamond",
			MaxStack:    64,
			Durability:  0,
			Weight:      0.5,
			Value:       50,
			Tags:        []string{"gem", "precious", "material"},
		},
		
		// Tools
		{
			ID:          "wooden_pickaxe",
			Name:        "Wooden Pickaxe",
			Type:        ItemTypeTool,
			Description: "A basic wooden pickaxe",
			MaxStack:    1,
			Durability:  60,
			Weight:      2.0,
			Value:       5,
			Tags:        []string{"tool", "pickaxe", "wooden"},
			Properties: map[string]interface{}{
				"mining_power": 1,
				"mining_speed": 1.0,
			},
		},
		{
			ID:          "stone_pickaxe",
			Name:        "Stone Pickaxe",
			Type:        ItemTypeTool,
			Description: "A stone pickaxe",
			MaxStack:    1,
			Durability:  132,
			Weight:      3.0,
			Value:       10,
			Tags:        []string{"tool", "pickaxe", "stone"},
			Properties: map[string]interface{}{
				"mining_power": 2,
				"mining_speed": 1.2,
			},
		},
		{
			ID:          "iron_pickaxe",
			Name:        "Iron Pickaxe",
			Type:        ItemTypeTool,
			Description: "An iron pickaxe",
			MaxStack:    1,
			Durability:  251,
			Weight:      3.5,
			Value:       25,
			Tags:        []string{"tool", "pickaxe", "iron"},
			Properties: map[string]interface{}{
				"mining_power": 3,
				"mining_speed": 1.5,
			},
		},
		{
			ID:          "diamond_pickaxe",
			Name:        "Diamond Pickaxe",
			Type:        ItemTypeTool,
			Description: "A diamond pickaxe",
			MaxStack:    1,
			Durability:  1561,
			Weight:      4.0,
			Value:       100,
			Tags:        []string{"tool", "pickaxe", "diamond"},
			Properties: map[string]interface{}{
				"mining_power": 4,
				"mining_speed": 2.0,
			},
		},
		
		// Weapons
		{
			ID:          "wooden_sword",
			Name:        "Wooden Sword",
			Type:        ItemTypeWeapon,
			Description: "A basic wooden sword",
			MaxStack:    1,
			Durability:  60,
			Weight:      2.0,
			Value:       5,
			Tags:        []string{"weapon", "sword", "wooden"},
			Properties: map[string]interface{}{
				"damage": 4,
				"attack_speed": 1.6,
			},
		},
		{
			ID:          "iron_sword",
			Name:        "Iron Sword",
			Type:        ItemTypeWeapon,
			Description: "An iron sword",
			MaxStack:    1,
			Durability:  251,
			Weight:      3.0,
			Value:       25,
			Tags:        []string{"weapon", "sword", "iron"},
			Properties: map[string]interface{}{
				"damage": 6,
				"attack_speed": 1.6,
			},
		},
		
		// Food
		{
			ID:          "apple",
			Name:        "Apple",
			Type:        ItemTypeFood,
			Description: "A juicy apple",
			MaxStack:    64,
			Durability:  0,
			Weight:      0.2,
			Value:       2,
			Tags:        []string{"food", "fruit"},
			Properties: map[string]interface{}{
				"hunger_restoration": 4,
				"health_restoration": 2,
			},
		},
		{
			ID:          "bread",
			Name:        "Bread",
			Type:        ItemTypeFood,
			Description: "A loaf of bread",
			MaxStack:    64,
			Durability:  0,
			Weight:      0.5,
			Value:       3,
			Tags:        []string{"food", "prepared"},
			Properties: map[string]interface{}{
				"hunger_restoration": 5,
				"health_restoration": 0,
			},
		},
		
		// Armor
		{
			ID:          "leather_helmet",
			Name:        "Leather Helmet",
			Type:        ItemTypeArmor,
			Description: "A leather helmet",
			MaxStack:    1,
			Durability:  55,
			Weight:      1.0,
			Value:       5,
			Tags:        []string{"armor", "helmet", "leather"},
			Properties: map[string]interface{}{
				"protection": 1,
				"armor_slot": "head",
			},
		},
		{
			ID:          "iron_chestplate",
			Name:        "Iron Chestplate",
			Type:        ItemTypeArmor,
			Description: "An iron chestplate",
			MaxStack:    1,
			Durability:  240,
			Weight:      5.0,
			Value:       30,
			Tags:        []string{"armor", "chestplate", "iron"},
			Properties: map[string]interface{}{
				"protection": 6,
				"armor_slot": "chest",
			},
		},
	}
	
	for _, item := range defaultItems {
		ir.items[item.ID] = item
		ir.byType[item.Type] = append(ir.byType[item.Type], item)
	}
}

// Global item registry instance
var GlobalItemRegistry = NewItemRegistry()

// GetGlobalItemRegistry returns the global item registry
func GetGlobalItemRegistry() *ItemRegistry {
	return GlobalItemRegistry
}

// Errors
var (
	ErrItemAlreadyExists = &ItemError{Message: "item already exists"}
	ErrItemNotFound     = &ItemError{Message: "item not found"}
	ErrInvalidItemType  = &ItemError{Message: "invalid item type"}
	ErrStackFull       = &ItemError{Message: "stack is full"}
	ErrCannotStack     = &ItemError{Message: "items cannot be stacked"}
)

// ItemError represents an item-related error
type ItemError struct {
	Message string
}

func (e *ItemError) Error() string {
	return e.Message
}
